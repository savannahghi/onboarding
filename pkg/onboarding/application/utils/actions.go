package utils

import (
	"context"
	"sort"

	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/profileutils"
)

//CheckUserHasPermission takes in the user roles and a permission and verifies that the user
//has required permissions
func CheckUserHasPermission(roles []profileutils.Role, permission profileutils.Permission) bool {
	scopes := []string{}
	for _, role := range roles {
		// only add scopes of active roles
		if role.Active {
			scopes = append(scopes, role.Scopes...)
		}
	}

	for _, scope := range scopes {
		if permission.Scope == scope {
			return true
		}
	}
	return false
}

//RemoveDuplicateStrings removes duplicate strings from a list of strings
func RemoveDuplicateStrings(strings []string) []string {
	mapped := make(map[string]string)
	cleaned := []string{}

	for _, v := range strings {
		_, inMap := mapped[v]
		if !inMap {
			mapped[v] = v
			cleaned = append(cleaned, v)
		}
	}

	return cleaned
}

//GetUserPermissions  returns all the scopes of user permissions
func GetUserPermissions(roles []profileutils.Role) []string {
	scopes := []string{}
	for _, role := range roles {
		// only add scopes of active roles
		if role.Active {
			scopes = append(scopes, role.Scopes...)
		}
	}
	cleaned := RemoveDuplicateStrings(scopes)

	return cleaned
}

// GetUserNavigationActions returns a sorted primary and secondary user navigation actions
func GetUserNavigationActions(
	ctx context.Context,
	user profileutils.UserProfile,
	roles []profileutils.Role,
	actions []domain.NavigationAction,
) (*dto.GroupedNavigationActions, error) {
	//all user actions
	userNavigationActions := []domain.NavigationAction{}

	for _, action := range actions {
		if action.RequiredPermission == nil || CheckUserHasPermission(roles, *action.RequiredPermission) {
			//  check for favorite navigation actions
			if IsFavNavAction(&user, action.Title) {
				action.Favorite = true
			}

			userNavigationActions = append(userNavigationActions, action)
		}
	}

	groupNested := GroupNested(userNavigationActions)
	primary, secondary := GroupPriority(groupNested)

	navigationActions := &dto.GroupedNavigationActions{
		Primary:   primary,
		Secondary: secondary,
	}
	return navigationActions, nil
}

// GroupNested groups navigation actions into parents (non nested actions) and children(nested actions)
func GroupNested(
	actions []domain.NavigationAction,
) []domain.NavigationAction {

	// Array of all parent actions i.e can have nested actions
	nonNested := []domain.NavigationAction{}
	for _, action := range actions {
		if !action.HasParent {
			nonNested = append(nonNested, action)
		}
	}

	// An array of properly grouped actions
	// The parent action has the nested actions
	grouped := []domain.NavigationAction{}
	for _, parent := range nonNested {
		// add the nested actions if any
		for _, action := range actions {
			if action.HasParent && action.Group == parent.Group {
				parent.Nested = append(parent.Nested, action)
			}
		}

		//add only the navigation actions that either has onTapRoute or has nested actions
		if len(parent.Nested) > 0 || parent.OnTapRoute != "" {
			// for an action with a single nested nested action
			// treat the nested action as a standalone non nested action
			if len(parent.Nested) == 1 {
				a := parent.Nested[0].(domain.NavigationAction)
				a.Icon = parent.Icon
				grouped = append(grouped, a)

			} else {
				grouped = append(grouped, parent)
			}
		}
	}

	return grouped
}

// GroupPriority groups navigation actions into primary and secondary actions
func GroupPriority(
	actions []domain.NavigationAction,
) (primary, secondary []domain.NavigationAction) {

	// sort actions based on priority using the sequence number
	// uses the inbuilt go sorting functionality
	// https://pkg.go.dev/sort#SliceStable
	sort.SliceStable(actions, func(i, j int) bool {
		return actions[i].SequenceNumber < actions[j].SequenceNumber
	})

	// nonNested has actions without nested actions
	// only non nested actions qualify as primary actions
	nonNested := []domain.NavigationAction{}
	nested := []domain.NavigationAction{}
	for _, action := range actions {
		if len(action.Nested) == 0 {
			nonNested = append(nonNested, action)
		} else {
			nested = append(nested, action)
		}
	}

	// primary actions constraints:
	//	- minimum of 2 i.e home and/or help
	// 	- maximum of 4
	primary = []domain.NavigationAction{}

	secondary = []domain.NavigationAction{}

	// all non nested actions can be primary actions
	if len(nonNested) <= 4 {
		// add the non nested actions to primary
		primary = append(primary, nonNested...)

		// add the rest to secondary
		secondary = append(secondary, nested...)

	} else {

		// add the first four to primary
		primary = nonNested[0:4]
		// add the rest to secondary
		secondary = nonNested[4:]

		// all nested actions to secondary
		secondary = append(secondary, nested...)

	}

	// sort the primary and secondary actions based on priority again
	// this is a precautionary step since all actions were sorted before
	sort.SliceStable(primary, func(i, j int) bool {
		return primary[i].SequenceNumber < primary[j].SequenceNumber
	})

	sort.SliceStable(secondary, func(i, j int) bool {
		return secondary[i].SequenceNumber < secondary[j].SequenceNumber
	})

	return primary, secondary
}
