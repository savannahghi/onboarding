// Package interactor represent reusable chunks of code that abstract
// logic from presenters while simplifying your app and making future changes effortless.
package interactor

import (
	"github.com/savannahghi/onboarding/pkg/onboarding/application/extension"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure"
	"github.com/savannahghi/onboarding/pkg/onboarding/usecases"
	"github.com/savannahghi/onboarding/pkg/onboarding/usecases/admin"
)

// Usecases is an interface that combines of all usescases
type Usecases interface {
	usecases.ProfileUseCase
	usecases.SignUpUseCases
	usecases.LoginUseCases
	usecases.SurveyUseCases
	usecases.UserPINUseCases
	admin.Usecase
	usecases.RoleUseCase
	usecases.PermissionUseCase
}

// Interactor is an implementation of the usecases interface
type Interactor struct {
	usecases.LoginUseCases
	usecases.ProfileUseCase
	usecases.RoleUseCase
	usecases.SignUpUseCases
	usecases.SurveyUseCases
	usecases.UserPINUseCases
	admin.Usecase
	usecases.PermissionUseCase
}

// NewUsecasesInteractor initializes a new usecases interactor
func NewUsecasesInteractor(
	infrastructure infrastructure.Infrastructure,
	baseExtension extension.BaseExtension,
	pinsExtension extension.PINExtension) Usecases {

	profile := usecases.NewProfileUseCase(infrastructure, baseExtension)
	login := usecases.NewLoginUseCases(infrastructure, profile, baseExtension, pinsExtension)
	roles := usecases.NewRoleUseCases(infrastructure, baseExtension)
	pins := usecases.NewUserPinUseCase(infrastructure, profile, baseExtension, pinsExtension)
	signup := usecases.NewSignUpUseCases(infrastructure, profile, pins, baseExtension)
	surveys := usecases.NewSurveyUseCases(infrastructure, baseExtension)
	services := admin.NewService(baseExtension)
	permissions := usecases.NewPermissionUseCases(infrastructure, baseExtension)

	impl := &Interactor{
		login,
		profile,
		roles,
		signup,
		surveys,
		pins,
		services,
		permissions,
	}

	return impl
}
