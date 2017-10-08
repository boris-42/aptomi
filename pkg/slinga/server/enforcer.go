package server

import (
	"github.com/Aptomi/aptomi/pkg/slinga/db"
	//"github.com/Aptomi/aptomi/pkg/slinga/engine/apply"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/diff"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/slinga/event"
	"github.com/Aptomi/aptomi/pkg/slinga/external"
	"github.com/Aptomi/aptomi/pkg/slinga/external/secrets"
	"github.com/Aptomi/aptomi/pkg/slinga/external/users"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"github.com/Aptomi/aptomi/pkg/slinga/server/store"
	//"github.com/Aptomi/aptomi/pkg/slinga/visualization"
	"github.com/Aptomi/aptomi/pkg/slinga/engine/apply"
	log "github.com/Sirupsen/logrus"
)

type Enforcer struct {
	store store.ServerStore
}

func NewEnforcer(store store.ServerStore) *Enforcer {
	return &Enforcer{store}
}

func (e *Enforcer) Enforce() {
	policy, err := e.store.GetPolicy(object.LastGen)
	if err != nil {
		// todo
		log.Panicf("Error while getting last policy: %s", err)
	}

	// todo empty state temporarily
	actualState := resolve.NewPolicyResolution()

	externalData := external.NewData(
		users.NewUserLoaderFromLDAP(db.GetAptomiPolicyDir()),
		secrets.NewSecretLoaderFromDir(db.GetAptomiPolicyDir()),
	)

	resolver := resolve.NewPolicyResolver(policy, externalData)
	desiredState, eventLog, err := resolver.ResolveAllDependencies()
	if err != nil {
		log.Panicf("Cannot resolve policy: %v %v %v", err, desiredState, actualState)
	}

	eventLog.Save(&event.HookStdout{})

	revision, err := e.store.NextRevision()
	if err != nil {
		log.Panicf("Unable to get next revision", err)
	}

	stateDiff := diff.NewPolicyResolutionDiff(desiredState, actualState, revision.GetGeneration())

	if !stateDiff.IsChanged() {
		log.Infof("No changes")
		return
	}
	log.Infof("Changes")

	// todo generate diagrams
	//prefDiagram := visualization.NewDiagram(actualPolicy, actualState, externalData)
	//newDiagram := visualization.NewDiagram(policy, desiredState, externalData)
	//deltaDiagram := visualization.NewDiagramDelta(policy, desiredState, actualPolicy, actualState, externalData)
	//visualization.CreateImage(...) for all diagrams

	applier := apply.NewEngineApply(policy, desiredState, actualPolicy, actualState, e.store.ActualStateUpdater(), externalData, plugins, stateDiff.Actions)
	resolution, eventLog, err := applier.Apply()
	if err != nil {

	}

	eventLog.Save(&event.HookStdout{})

}

/*
func (ctl *RevisionControllerImpl) CheckState() error {
	policy, err := ctl.policyCtl.GetPolicy(object.LastGen)

	// Background applier
	// [1] Calculate current desired state (run resolver // resolver.ResolveAllDependencies())
	// 	 1. Use latest policy version and latest external data
	// [2] Load current actual state
	// [3] Compare actual and desired state, calculate diff and actions
	//   1. Note(!): actual state will not have an associated policy
	//               do not use Prev.Policy in diff or apply
	// [4] If hasChanges =>
	// 				new Revision,
	//				attach resolution event log to the revision (may be attach to RevisionSummary)
	// 		else return no new revision needed
	// [5] Applies executes all actions and updates actual state, if/as needed
	//   1. Once action has been executed, save its status and event log to DB
	// [6] Mark revision as "OK" if all actions were completed without errors
	// [7] Keep RevisionSummary
	// 	 1. process text policy diff, add to RevisionSummary
	// 	 1. generate new charts - NewPolicyVisualization, add to RevisionSummary
	// 	 1. save text diff for component instances into RevisionSummary

	// Load the previous usage state (for now it's just empty), it's for now ActualState
	prevState := resolve.NewPolicyResolution()

	externalData := external.NewData(
		users.NewUserLoaderFromLDAP(db.GetAptomiPolicyDir()),
		secrets.NewSecretLoaderFromDir(db.GetAptomiPolicyDir()),
	)
	resolver := resolve.NewPolicyResolver(policy, externalData)
	resolution, eventLog, err := resolver.ResolveAllDependencies()
	if err != nil {
		eventLog.Save(&event.HookStdout{})
		log.Panicf("Cannot resolve policy: %v %v %v", err, resolution, prevState)
	}
	eventLog.Save(&event.HookStdout{})

	fmt.Println("Success")
	fmt.Println("Components:", len(resolution.ComponentInstanceMap))

	return nil
}

/*
	PolicyResolver should return PolicyResolution
	Remove revision from the engine
	Revision in controller package (created by Sergey) = policyResolution + actions
*/

/*
			// Get loader for external users
			userLoader := NewAptomiUserLoader()

			// Load the previous usage state
			prevState := resolve.LoadRevision()

			policy := ... NewPolicy()

			resolver := resolve.NewPolicyResolver(policy, userLoader)
			nextState, err := resolver.ResolveAllDependencies()
			if err != nil {
				log.Panicf("Cannot resolve policy: %v", err)
			}

			// Process differences
			diff := NewRevisionDiff(nextState, prevState)
			diff.AlterDifference(full)
			diff.StoreDiffAsText(verbose)

			// Print on screen
			fmt.Print(diff.DiffAsText)

			// Generate pictures (see new API :)

			// Save new resolved state in the last run directory
			resolver.SavePolicyResolution() -> this called before:

					revision := NewRevision(resolver.policy, resolver.resolution, resolver.userLoader)
					revision.Save()

					// Save log
					hook := &HookBoltDB{}
					resolver.eventLog.Save(hook)


	/*
			///////////////////// APPLY

			// Apply changes (if emulateDeployment == true --> we set noop to skip deployment part)
			apply := NewEngineApply(diff)
			if !(noop || emulateDeployment) {
				err := apply.Apply()
				apply.SaveLog()
				if err != nil {
					log.Panicf("Cannot apply policy: %v", err)
				}
			}

			// If everything is successful, then increment revision and save run
			// if emulateDeployment == true --> we set noop to false to write state on disk)
			revision := GetLastRevision(GetAptomiBaseDir())
			diff.ProcessSuccessfulExecution(revision, newrevision, noop && !emulateDeployment)
*/
