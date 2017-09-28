package language

import (
	"github.com/Aptomi/aptomi/pkg/slinga/language/expression"
	"github.com/Aptomi/aptomi/pkg/slinga/object"
	"sort"
	"sync"
)

var RuleObject = &object.Info{
	Kind:        "rule",
	Versioned: true,
	Constructor: func() object.Base { return &Rule{} },
}

// Rule is a rule
type Rule struct {
	Metadata

	Weight   int
	Criteria *Criteria
	Actions  *RuleActions
}

// RuleActions is a set of actions performed by rule
type RuleActions struct {
	ChangeLabels ChangeLabelsAction `yaml:"change-labels"`
	Dependency   DependencyAction
	Ingress      IngressAction
}

// Matches returns if a rule matches
func (rule *Rule) Matches(params *expression.ExpressionParameters, cache *expression.ExpressionCache) (bool, error) {
	if rule.Criteria == nil {
		return true, nil
	}
	return rule.Criteria.allows(params, cache)
}

// GlobalRules is a list of global rules
type GlobalRules struct {
	Rules []*Rule

	once        sync.Once
	rulesSorted []*Rule // lazily initialized value
}

// NewGlobalRules creates and initializes a new empty list of global rules
func NewGlobalRules() *GlobalRules {
	return &GlobalRules{}
}

func (globalRules *GlobalRules) addRule(rule *Rule) {
	globalRules.Rules = append(globalRules.Rules, rule)
}

func (globalRules *GlobalRules) GetRulesSortedByWeight() []*Rule {
	globalRules.once.Do(func() {
		globalRules.rulesSorted = append(globalRules.rulesSorted, globalRules.Rules...)
		sort.Sort(globalRules)
	})
	return globalRules.rulesSorted
}

func (globalRules *GlobalRules) Len() int {
	return len(globalRules.rulesSorted)
}

func (globalRules *GlobalRules) Less(i, j int) bool {
	return globalRules.rulesSorted[i].Weight < globalRules.rulesSorted[j].Weight
}

func (globalRules *GlobalRules) Swap(i, j int) {
	globalRules.rulesSorted[i], globalRules.rulesSorted[j] = globalRules.rulesSorted[j], globalRules.rulesSorted[i]
}
