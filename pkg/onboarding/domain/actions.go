package domain

import (
	"github.com/savannahghi/profileutils"
)

// Icon links for navactions
const (
	// StaticBase is the default path at which static assets are hosted
	StaticBase = "https://assets.healthcloud.co.ke"

	RoleNavActionIcon     = StaticBase + "/actions/roles_navaction.png"
	ConsumerNavActionIcon = StaticBase + "/actions/consumer_navaction.png"
	HelpNavActionIcon     = StaticBase + "/actions/help_navaction.png"
	HomeNavActionIcon     = StaticBase + "/actions/home_navaction.png"
	PartnerNavActionIcon  = StaticBase + "/actions/partner_navaction.png"
	PatientNavActionIcon  = StaticBase + "/actions/patient_navaction.png"
	RequestNavActionIcon  = StaticBase + "/actions/request_navaction.png"
)

// On Tap Routes
const (
	HomeRoute                  = "/home"
	PatientRegistrationRoute   = "/addPatient"
	PatientIdentificationRoute = "/patients"
	GetHelpRouteRoute          = "/helpCenter"
	RequestsRoute              = "/admin"
	RoleViewRoute              = "/viewCreatedRolesPage"
	RoleCreationRoute          = "/createRoleStepOne"
	RoleAssignmentRoute        = "/bewellUserIdentification"
)

// Navigation actions
const (
	HomeNavActionTitle       = "Home"
	HomeNavActionDescription = "Home Navigation action"

	HelpNavActionTitle       = "Help"
	HelpNavActionDescription = "Help Navigation action"

	RoleNavActionTitle      = "Role Management"
	RoleViewActionTitle     = "View Roles"
	RoleCreationActionTitle = "Create Role"
	RoleAssignActionTitle   = "Assign Role"

	PatientNavActionTitle            = "Patients"
	PatientNavActionDescription      = "Patient Navigation action"
	PatientRegistrationActionTitle   = "Register Patient"
	PatientIdentificationActionTitle = "Search Patient"

	RequestsNavActionTitle       = "Requests"
	RequestsNavActionDescription = "Requests Navigation action"

	ConsumerNavActionTitle       = "Consumers"
	ConsumerNavActionDescription = "Consumer Navigation action"

	PartnerNavActionTitle       = "Partners"
	PartnerNavActionDescription = "Partner Navigation action"
)

const (
	//HomeGroup groups all actions under the home resource
	HomeGroup NavigationGroup = "home"

	//RoleGroup groups all actions under the role resource
	RoleGroup NavigationGroup = "role"

	//HelpGroup groups all actions under the help resource
	HelpGroup NavigationGroup = "help"

	//KYCGroup groups all actions under the kyc resource
	KYCGroup NavigationGroup = "kyc"

	//PatientGroup groups all actions under the patient resource
	PatientGroup NavigationGroup = "patient"

	//PartnerGroup groups all actions under the partner resource
	PartnerGroup NavigationGroup = "partner"

	//RolesGroup groups all actions under the role resource
	RolesGroup NavigationGroup = "role"

	//ConsumerGroup groups all actions under the consumer resource
	ConsumerGroup NavigationGroup = "consumer"
)

// Determines the sequence number of a navigation action
// Order of the constants matters!!
const (
	HomeNavActionSequence = iota + 1

	RoleNavActionSequence
	RoleCreationNavActionSequence
	RoleViewingNavActionSequence
	RoleAssignNavActionSequence

	RequestsNavActionSequence

	PartnerNavactionSequence

	ConsumerNavactionSequence

	PatientNavActionSequence
	PatientSearchNavActionSequence
	PatientRegistrationNavActionSequence

	HelpNavActionSequence
)

// the structure and definition of all navigation actions
var (
	// HomeNavAction is the primary home button
	HomeNavAction = NavigationAction{
		Group:              HomeGroup,
		Title:              HomeNavActionTitle,
		OnTapRoute:         HomeRoute,
		Icon:               HomeNavActionIcon,
		RequiredPermission: nil,
		SequenceNumber:     HomeNavActionSequence,
	}

	// HelpNavAction navigation action to help and FAQs page
	HelpNavAction = NavigationAction{
		Group:              HelpGroup,
		Title:              HelpNavActionTitle,
		OnTapRoute:         GetHelpRouteRoute,
		Icon:               HelpNavActionIcon,
		RequiredPermission: nil,
		SequenceNumber:     HelpNavActionSequence,
	}
)

var (

	// KYCNavActions is the navigation acction to KYC processing
	KYCNavActions = NavigationAction{
		Group:              KYCGroup,
		Title:              RequestsNavActionTitle,
		OnTapRoute:         RequestsRoute,
		Icon:               RequestNavActionIcon,
		RequiredPermission: &profileutils.CanProcessKYC,
		SequenceNumber:     RequestsNavActionSequence,
	}
)

var (
	//PartnerNavActions is the navigation actions to partner management
	PartnerNavActions = NavigationAction{
		Group:              PartnerGroup,
		Title:              PartnerNavActionTitle,
		Icon:               PartnerNavActionIcon,
		RequiredPermission: &profileutils.CanViewPartner,
		SequenceNumber:     PartnerNavactionSequence,
	}
)

var (
	//ConsumerNavActions is the navigation actions to consumer management
	ConsumerNavActions = NavigationAction{
		Group:              ConsumerGroup,
		Title:              ConsumerNavActionTitle,
		Icon:               ConsumerNavActionIcon,
		RequiredPermission: &profileutils.CanViewConsumers,
		SequenceNumber:     ConsumerNavactionSequence,
	}
)

var (
	//RoleNavActions this is the parent navigation action for role resource
	// it has nested navigation actions below
	RoleNavActions = NavigationAction{
		Group:              RoleGroup,
		Title:              RoleNavActionTitle,
		Icon:               RoleNavActionIcon,
		RequiredPermission: nil,
		SequenceNumber:     RoleNavActionSequence,
	}

	//RoleCreationNavAction a child of the RoleNavActions
	RoleCreationNavAction = NavigationAction{
		Group:              RoleGroup,
		Title:              RoleCreationActionTitle,
		OnTapRoute:         RoleCreationRoute,
		RequiredPermission: &profileutils.CanCreateRole,
		HasParent:          true,
		SequenceNumber:     RoleCreationNavActionSequence,
	}

	//RoleViewNavAction a child of the RoleNavActions
	RoleViewNavAction = NavigationAction{
		Group:              RoleGroup,
		Title:              RoleViewActionTitle,
		OnTapRoute:         RoleViewRoute,
		RequiredPermission: &profileutils.CanViewRole,
		HasParent:          true,
		SequenceNumber:     RoleViewingNavActionSequence,
	}

	//RoleAssignNavAction a child of the RoleNavActions
	RoleAssignNavAction = NavigationAction{
		Group:              RoleGroup,
		Title:              RoleAssignActionTitle,
		OnTapRoute:         RoleAssignmentRoute,
		RequiredPermission: &profileutils.CanAssignRole,
		HasParent:          true,
		SequenceNumber:     RoleAssignNavActionSequence,
	}
)

var (
	//PatientNavActions this is the parent navigation action for patient resource
	// it has nested navigation actions below
	PatientNavActions = NavigationAction{
		Group:              PatientGroup,
		Title:              PatientNavActionTitle,
		Icon:               PatientNavActionIcon,
		RequiredPermission: nil,
		SequenceNumber:     PatientNavActionSequence,
	}

	//PatientRegistrationNavAction a child of the PatientNavActions
	PatientRegistrationNavAction = NavigationAction{
		Group:              PatientGroup,
		Title:              PatientRegistrationActionTitle,
		OnTapRoute:         PatientRegistrationRoute,
		RequiredPermission: &profileutils.CanCreatePatient,
		HasParent:          true,
		SequenceNumber:     PatientRegistrationNavActionSequence,
	}

	//PatientIdentificationNavAction a child of the PatientNavActions
	PatientIdentificationNavAction = NavigationAction{
		Group:              PatientGroup,
		Title:              PatientIdentificationActionTitle,
		OnTapRoute:         PatientIdentificationRoute,
		RequiredPermission: &profileutils.CanIdentifyPatient,
		HasParent:          true,
		SequenceNumber:     PatientSearchNavActionSequence,
	}
)

// AllNavigationActions is a grouping of all navigation actions
var AllNavigationActions = []NavigationAction{
	HomeNavAction, HelpNavAction,

	KYCNavActions, PartnerNavActions, ConsumerNavActions,

	PatientNavActions, PatientRegistrationNavAction, PatientIdentificationNavAction,

	RoleNavActions, RoleCreationNavAction, RoleViewNavAction,
}
