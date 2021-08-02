package domain

import (
	"context"
	"time"

	"github.com/savannahghi/onboarding/pkg/onboarding/application/common"
	"github.com/savannahghi/profileutils"
)

var (
	// TimeLocation ...
	TimeLocation, _ = time.LoadLocation("Africa/Nairobi")

	// TimeFormatStr date time string format
	TimeFormatStr = "2006-01-02T15:04:05+03:00"

	// Repo the env to identify which repo to use
	Repo = "REPOSITORY"

	//FirebaseRepository is the value of the env when using firebase
	FirebaseRepository = "firebase"

	//PostgresRepository is the value of the env when using postgres
	PostgresRepository = "postgres"
)

const (
	//HomeGroup groups all actions under the home resource
	HomeGroup NavigationGroup = "home"

	//RoleGroup groups all actions under the role resource
	RoleGroup NavigationGroup = "role"

	//HelpGroup groups all actions under the help resource
	HelpGroup NavigationGroup = "help"

	//AgentGroup groups all actions under the agent resource
	AgentGroup NavigationGroup = "agents"

	//KYCGroup groups all actions under the kyc resource
	KYCGroup NavigationGroup = "kyc"

	//EmployeeGroup groups all actions under the employees resource
	EmployeeGroup NavigationGroup = "employees"

	//CoversGroup groups all actions under the covers resource
	CoversGroup NavigationGroup = "covers"

	//PatientGroup groups all actions under the patient resource
	PatientGroup NavigationGroup = "patient"

	//PartnerGroup groups all actions under the partner resource
	PartnerGroup NavigationGroup = "partner"

	//RolesGroup groups all actions under the role resource
	RolesGroup NavigationGroup = "role"

	//ConsumerGroup groups all actions under the consumer resource
	ConsumerGroup NavigationGroup = "consumer"
)

// the structure and definition of all navigation actions
var (
	// HomeNavAction is the primary home button
	HomeNavAction = NavigationAction{
		Group:              HomeGroup,
		Title:              common.HomeNavActionTitle,
		OnTapRoute:         common.HomeRoute,
		Icon:               common.HomeNavActionURL,
		IsHighPriority:     true,
		RequiredPermission: nil,
	}

	// HelpNavAction navigation action to help and FAQs page
	HelpNavAction = NavigationAction{
		Group:              HelpGroup,
		Title:              common.HelpNavActionTitle,
		OnTapRoute:         common.GetHelpRouteRoute,
		Icon:               common.HelpNavActionURL,
		RequiredPermission: nil,
	}

	// KYCNavActions is the navigation acction to KYC processing
	KYCNavActions = NavigationAction{
		Group:              KYCGroup,
		Title:              common.RequestsNavActionTitle,
		OnTapRoute:         common.RequestsRoute,
		Icon:               common.RequestNavActionURL,
		IsHighPriority:     true,
		RequiredPermission: &profileutils.CanProcessKYC,
	}

	//PartnerNavActions is the navigation actions to partner management
	PartnerNavActions = NavigationAction{
		Group: PartnerGroup,
		Title: common.PartnerNavActionTitle,
		// Not provided yet
		OnTapRoute:         "",
		Icon:               common.PartnerNavActionURL,
		IsHighPriority:     true,
		RequiredPermission: &profileutils.CanViewPartner,
	}

	//ConsumerNavActions is the navigation actions to consumer management
	ConsumerNavActions = NavigationAction{
		Group: ConsumerGroup,
		Title: common.ConsumerNavActionTitle,
		// Not provided yet
		OnTapRoute:         "",
		Icon:               common.ConsumerNavActionURL,
		RequiredPermission: &profileutils.CanViewConsumers,
	}

	//RoleNavActions this is the parent navigation action for role resource
	// it has nested navigation actions below
	RoleNavActions = NavigationAction{
		Group:              RoleGroup,
		Title:              common.RoleNavActionTitle,
		Icon:               common.RoleNavActionURL,
		RequiredPermission: &profileutils.CanViewRole,
	}

	//RoleCreationNavActions a child of the RoleNavActions
	RoleCreationNavActions = NavigationAction{
		Group:              RoleGroup,
		Title:              common.RoleCreationActionTitle,
		OnTapRoute:         common.RoleCreationRoute,
		RequiredPermission: &profileutils.CanCreateRole,
		HasParent:          true,
	}

	//RoleViewNavActions a child of the RoleNavActions
	RoleViewNavActions = NavigationAction{
		Group:              RoleGroup,
		Title:              common.RoleViewActionTitle,
		OnTapRoute:         common.RoleViewRoute,
		RequiredPermission: &profileutils.CanViewRole,
		HasParent:          true,
	}

	//AgentNavActions this is the parent navigation action for agent resource
	// it has nested navigation actions below
	AgentNavActions = NavigationAction{
		Group:              AgentGroup,
		Title:              common.AgentNavActionTitle,
		Icon:               common.AgentNavActionURL,
		RequiredPermission: &profileutils.CanViewAgent,
	}

	//AgentRegistrationNavActions a child of the AgentNavActions
	AgentRegistrationNavActions = NavigationAction{
		Group:              AgentGroup,
		Title:              common.AgentRegistrationActionTitle,
		OnTapRoute:         common.AgentRegistrationRoute,
		RequiredPermission: &profileutils.CanRegisterAgent,
		HasParent:          true,
	}

	//AgentidentificationNavActions a child of the AgentNavActions
	AgentidentificationNavActions = NavigationAction{
		Group:              AgentGroup,
		Title:              common.AgentIdentificationActionTitle,
		OnTapRoute:         common.AgentIdentificationRoute,
		RequiredPermission: &profileutils.CanIdentifyAgent,
		HasParent:          true,
	}

	//PatientNavActions this is the parent navigation action for patient resource
	// it has nested navigation actions below
	PatientNavActions = NavigationAction{
		Group:              PatientGroup,
		Title:              common.PatientNavActionTitle,
		Icon:               common.PatientNavActionURL,
		RequiredPermission: &profileutils.CanViewPatient,
	}

	//PatientRegistrationNavActions a child of the PatientNavActions
	PatientRegistrationNavActions = NavigationAction{
		Group:              PatientGroup,
		Title:              common.PatientRegistrationActionTitle,
		OnTapRoute:         common.PatientRegistrationRoute,
		RequiredPermission: &profileutils.CanCreatePatient,
		HasParent:          true,
	}

	//PatientIdentificationNavActions a child of the PatientNavActions
	PatientIdentificationNavActions = NavigationAction{
		Group:              PatientGroup,
		Title:              common.PatientIdentificationActionTitle,
		OnTapRoute:         common.PatientIdentificationRoute,
		RequiredPermission: &profileutils.CanIdentifyPatient,
		HasParent:          true,
	}
)

// GetAllNavigationActions returns a list of all the above defined navigation actions
func GetAllNavigationActions(ctx context.Context) ([]NavigationAction, error) {
	return []NavigationAction{
		HomeNavAction,

		HelpNavAction,

		KYCNavActions,

		PartnerNavActions,

		ConsumerNavActions,

		AgentNavActions,
		AgentRegistrationNavActions,
		AgentidentificationNavActions,

		PatientNavActions,
		PatientRegistrationNavActions,
		PatientIdentificationNavActions,
	}, nil
}
