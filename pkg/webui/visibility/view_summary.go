package visibility

// SummaryView represents summary view that we show on the home page
type SummaryView struct {
	userID string
	//revision *resolve.Revision
}

// NewSummaryView creates a new SummaryView
func NewSummaryView(userID string) SummaryView {
	return SummaryView{
		userID: userID,
		//revision: resolve.LoadRevision(),
	}
}

// GetData returns data for a given view
func (view SummaryView) GetData() interface{} {
	return map[string]interface{}{
		"globalDependencies": view.getGlobalDependenciesData(),
		"globalRules":        view.getGlobalRulesData(),
		"servicesOwned":      view.getServicesOwned(),
		"servicesUsing":      view.getServicesUsing(),
	}
}

func (view SummaryView) getGlobalDependenciesData() interface{} {
	// table only exists for global ops people
	result := lineEntryList{}
	/*
		if !view.revision.UserLoader.LoadUserByName(view.userID).IsGlobalOps() {
			return result
		}
		for _, dependency := range view.revision.Policy.Dependencies.DependenciesByID {
			entry := lineEntry{
				"resolved":        dependency.Resolved,
				"userName":        view.revision.UserLoader.LoadUserByName(dependency.UserID).Name,
				"serviceName":     dependency.Service,
				"contextWithKeys": view.getResolvedContextNameByDep(dependency),
				"cluster":         view.getResolvedClusterNameByDep(dependency),
				"stats":           view.getDependencyStats(dependency),
				"dependencyId":    dependency.GetID(),
				"id":              view.revision.UserLoader.LoadUserByName(dependency.UserID).Name, // entries will be sorted by ID
			}
			result = append(result, entry)
		}
		sort.Sort(result)
	*/
	return result
}

func (view SummaryView) getGlobalRulesData() interface{} {
	// table only exists for global ops people
	result := lineEntryList{}
	/*
		if !view.revision.UserLoader.LoadUserByName(view.userID).IsGlobalOps() {
			return result
		}
		for _, ruleList := range view.revision.Policy.Rules.Rules {
			for _, rule := range ruleList {
				entry := lineEntry{
					"ruleName":   rule.Name,
					"ruleObject": rule.FilterServices,
					"appliedTo":  view.getRuleAppliedTo(rule),
					// currently we're only matching users by labels (for demo with rules w/o any other filters)
					"matchedUsers": view.getRuleMatchedUsers(rule),
					"conditions":   rule.DescribeConditions(),
					"actions":      rule.DescribeActions(),
					"id":           rule.Name, // entries will be sorted by ID
				}
				result = append(result, entry)
			}
		}
		sort.Sort(result)
	*/
	return result
}

func (view SummaryView) getServicesOwned() interface{} {
	result := lineEntryList{}
	/*
		for _, service := range view.revision.Policy.Services {
			if service.Owner == view.userID {
				// if I own this service, let's find all its instances
				instanceMap := make(map[string]bool)
				for key, instance := range view.revision.Resolution.ComponentInstanceMap {
					if instance {
						if instance.Key.ServiceName == service.Name && instance.Key.IsService() {
							instanceMap[key] = true
						}
					}
				}

				// Add info about every allocated instance
				for key := range instanceMap {
					instance := view.revision.Resolution.ComponentInstanceMap[key]
					entry := lineEntry{
						"serviceName":     service.Name,
						"contextWithKeys": view.getResolvedContextNameByInst(instance),
						"cluster":         view.getResolvedClusterNameByInst(instance),
						"stats":           view.getInstanceStats(instance),
						"id":              getWebIDByComponentKey(key), // entries will be sorted by ID
					}
					result = append(result, entry)
				}
			}
		}
		sort.Sort(result)
	*/
	return result
}

func (view SummaryView) getServicesUsing() interface{} {
	result := lineEntryList{}
	/*
		for _, dependency := range view.revision.Policy.Dependencies.DependenciesByID {
			if dependency.UserID == view.userID {
				entry := lineEntry{
					"resolved":        dependency.Resolved,
					"serviceName":     dependency.Service,
					"contextWithKeys": view.getResolvedContextNameByDep(dependency),
					"cluster":         view.getResolvedClusterNameByDep(dependency),
					"stats":           view.getDependencyStats(dependency),
					"dependencyId":    dependency.GetID(),
					"id":              dependency.GetID(), // entries will be sorted by ID
				}
				result = append(result, entry)
			}
		}
		sort.Sort(result)
	*/
	return result
}
