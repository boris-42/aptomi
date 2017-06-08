#!/bin/bash

set -eou pipefail

#################### Setup debug and colors

COLOR_GRAY='\033[0;37m'
COLOR_BLUE='\033[0;34m'
COLOR_RED='\033[0;31m'
COLOR_RESET='\033[0m'

function finish() {
    echo -e -n $COLOR_RESET
}
trap finish EXIT

echo -e -n $COLOR_GRAY

DEBUG=${DEBUG:-no}
if [ "yes" == "$DEBUG" ]; then
    set -x
fi

#################### Main section

function main() {
    if [ "$#" -ne "1" ]; then
        echo "ERROR: (gke-demo.sh) Too few arguments"
        echo "Usage: gke-demo.sh <up | status | down>"
        exit 1
    fi

    # TODO(slukjanov): should we load params from config file?
    # defaults
    k8s_version=1.5.7
    disk_size=100

    cluster_big_name=cluster-us-west
    cluster_small_name=cluster-us-east

    cluster_big_region=us-west1-c
    cluster_small_region=us-east1-c

    cluster_big_size=1
    cluster_small_size=1

    # see https://cloud.google.com/compute/pricing#standard_machine_types
    cluster_big_flavor=n1-standard-8
    cluster_small_flavor=n1-standard-2

    cluster_disk_size=100

    firewall_rules_name=demo-firewall-open-all
    firewall_rules="--allow tcp"

    demo_namespace=demo
    # end of defaults

    gcloud_check

    if [ "up" == "$1" ]; then
        gke_firewall_create $firewall_rules_name "$firewall_rules"

        # create big cluster
        gke_cluster_create $cluster_big_name $cluster_big_region $k8s_version $disk_size $cluster_big_flavor $cluster_big_size
        gke_cluster_kubectl_setup $cluster_big_name $cluster_big_region $demo_namespace
        k8s_alive $cluster_big_name
        helm_init $cluster_big_name

        # create small cluster
        gke_cluster_create $cluster_small_name $cluster_small_region $k8s_version $disk_size $cluster_small_flavor $cluster_small_size
        gke_cluster_kubectl_setup $cluster_small_name $cluster_small_region $demo_namespace
        k8s_alive $cluster_small_name
        helm_init $cluster_small_name
    elif [ "down" == "$1" ]; then
        gke_firewall_delete $firewall_rules_name

        # delete big cluster
        gke_cluster_delete $cluster_big_name $cluster_big_region
        gke_cluster_kubectl_cleanup $cluster_big_name $cluster_big_region

        # delete small cluster
        gke_cluster_delete $cluster_small_name $cluster_small_region
        gke_cluster_kubectl_cleanup $cluster_small_name $cluster_small_region
    elif [ "status" == "$1" ]; then
        k8s_alive $cluster_big_name
        helm_alive $cluster_big_name

        k8s_alive $cluster_small_name
        helm_alive $cluster_small_name
    else
        log "Unsupported command '$1'"
        exit 1
    fi
}

#################### Logging utils

function log() {
    set +x
    echo -e "$COLOR_BLUE[$(date +"%F %T")] gke-demo $COLOR_RED|$COLOR_RESET" $@$COLOR_GRAY
    if [ "yes" == "$DEBUG" ] ; then
        set -x
    fi
}

function cluster_log_name() {
    echo "'$name' ($zone)"
}

#################### Gcloud utils

function gcloud_check() {
    log "Gcloud verification"

    if ! gcloud auth list 2>/dev/null | grep "ACTIVE\$" ; then
        log "There is no active gcloud account"
        log "Run 'gcloud auth list' to get account name"
        log "If no entries run 'gcloud auth login' to setup account"
        log "Or run 'gcloud config set account <account>' to select account"
    fi

    if [ -z "$(gcloud config get-value project 2>/dev/null)" ]; then
        log "Gcloud project isn't set."
        log "You can find projects using 'gcloud projects list'"
        log "Use 'gcloud config set project <project_id>' to set project ID"
    fi

    if ! gke list 1>/dev/null ; then
        log "Can't get list of clusters, check permissions"
        exit 1
    fi

    if ! gcf list 1>/dev/null ; then
        log "Can't get firewall rules list, check permissions"
    fi
}

#################### Commands aliases

function gke() {
    gcloud container clusters $@
}

function gcf() {
    gcloud compute firewall-rules $@
}

#################### Cluster ops

function gke_cluster_exists() {
    name="$1"
    zone="$2"

    if gke describe $name --zone $zone 2>/dev/null | grep -q "^name: $name\$" ; then
        return 0
    else
        return 1
    fi
}

function gke_cluster_create() {
    name="$1"
    zone="$2"
    version="$3"
    disk_size="$4"
    machine_type="$5"
    num_nodes="$6"

    if gke_cluster_exists $name $zone ; then
        log "Cluster $(cluster_log_name) already exists, run cleanup first to re-create"
    else
        log "Creating cluster $(cluster_log_name)"

        gke create \
            $name \
            --cluster-version $version \
            --zone $zone \
            --disk-size $disk_size \
            --machine-type $machine_type \
            --num-nodes $num_nodes
    fi

    if gke describe $name --zone $zone | grep -q "^status: RUNNING\$" ; then
        log "Cluster $(cluster_log_name) is RUNNING"
    else
        log "Cluster $(cluster_log_name) isn't RUNNING, run cleanup and try again"
        exit 1
    fi
}

function gke_cluster_delete() {
    name="$1"
    zone="$2"

    if ! gke_cluster_exists $name $zone ; then
        log "Cluster $(cluster_log_name) doesn't exists"
    else
        log "Deleting cluster $(cluster_log_name)"

        if gke delete $name --zone $zone --quiet ; then
            log "Cluster $(cluster_log_name) deleted successfully"
        else
            log "Cluster $(cluster_log_name) deletion failed, try to re-run cleanup"
            exit 1
        fi
    fi
}

#################### Kubeconfig ops

function kcfg_user_of_context() {
    name="$1"
    kubectl config view -o=jsonpath="{.contexts[?(@.name==\"$name\")].context.user}"
}

function kcfg_cluster_of_context() {
    name="$1"
    kubectl config view -o=jsonpath="{.contexts[?(@.name==\"$name\")].context.cluster}"
}

function gke_cluster_kubectl_setup() {
    name="$1"
    zone="$2"
    namespace="$3"

    project="$(gcloud config get-value project 2>/dev/null)"

    if gke get-credentials $1 --zone $2 2>/dev/null ; then
        kcfg_name="gke_${project}_${zone}_${name}"
        context=$kcfg_name
        user="$(kcfg_user_of_context $context)"
        cluster="$(kcfg_cluster_of_context $context)"

        if [[ -z "$user" || -z "$cluster" ]]; then
            log "Failed getting user or cluster for installed context '$context'"
            exit 1
        fi

        kubectl config set-context $name --cluster=$cluster --user=$user --namespace=$namespace
        log "Kubeconfig context '$name' (alias for '$context') successfully added"
    else
        log "Can't get credentials for cluster $(cluster_log_name)"
        exit 1
    fi
}

function gke_cluster_kubectl_cleanup() {
    name="$1"
    zone="$2"

    project="$(gcloud config get-value project 2>/dev/null)"
    kcfg_name="gke_${project}_${zone}_${name}"

    log "Cleaning up Kubeconfig for $(cluster_log_name)"

    kubectl config unset users.$kcfg_name

    # TODO(slukjanov): Add error checks for it
    kubectl config delete-cluster $kcfg_name || true
    kubectl config delete-context $kcfg_name || true
    kubectl config delete-context $name || true
}

#################### GCE firewall ops

function gke_firewall_exists() {
    if gcf describe $name 2>/dev/null | grep -q "^name: $name\$" ; then
        return 0
    else
        return 1
    fi
}

function gke_firewall_create() {
    name="$1"
    rules="$2"

    if gke_firewall_exists ; then
        log "Firewall rules '$name' already exists, run cleanup first to re-create"
    else
        if gcf create $name $rules ; then
            log "Firewall rules '$name' successfully created"
        else
            log "Firewall rules '$name' creation failed"
            exit 1
        fi
    fi
}

function gke_firewall_delete() {
    name="$1"

    if ! gke_firewall_exists ; then
        log "Firewall rules '$name' doesn't exists"
    else
        log "Deleting firewall rules '$name'"

        if gcf delete $1 --quiet ; then
            log "Firewall rules '$name' deleted successfully"
        else
            log "Firewall rules '$name' deletion failed"
            exit 1
        fi
    fi
}

#################### K8s healt check

function k8s_alive() {
    name="$1"

    log "Verifying cluster $name"

    kubectl --context $name cluster-info
    kubectl --context $name get ns 1>/dev/null

    # magic number of running pods in kube-system namespace
    if [[ $(kubectl --context $name -n kube-system get pods | grep " Running " | wc -l) -ge 3 ]]; then
        return 0
    else
        log "Cluster $name seems not really alive"
    fi
}

#################### Helm utils

function helm_alive() {
    name="$1"

    if ! kubectl --context $name -n kube-system describe deploy tiller-deploy 2>/dev/null | grep -q "1 desired"; then
        log "Can't find Tiller deployment in cluster $name"
        return 1
    fi

    if ! helm --kube-context $1 list --all 2>/dev/null ; then
        log "Helm in cluster $name seems not really alive"
        return 1
    fi

    log "Helm in cluster $name seems alive"
    return 0
}

function helm_init() {
    name="$1"

    if ! helm_alive $name ; then
        if ! helm --kube-context $name init 2>/dev/null ; then
            log "Helm init failed in cluster $name"
            exit 1
        fi

        log "Waiting 10 seconds for Tiller to start"
        sleep 10

        if ! helm_alive $name ; then
            log "Helm isn't alive 10 seconds after running helm init, let's wait 60 seconds"
            sleep 60

            if ! helm_alive $name ; then
                log "Helm isn't alive 70 seconds after running helm init, fail"
                exit 1
            fi
        fi

        log "Helm in cluster $name successfully initialized"
    fi
}

#################### End

main $@
log "gke-demo.sh successfully finished"
