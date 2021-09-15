package usecases

import (
	"github.com/savannahghi/onboarding/pkg/onboarding/application/extension"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure"
	"github.com/savannahghi/onboarding/pkg/onboarding/usecases/admin"
)

// Interactor is an implementation of the usecases interface
type Interactor struct {
	LoginUseCases
	ProfileUseCase
	RoleUseCase
	SignUpUseCases
	SurveyUseCases
	UserPINUseCases
	admin.Usecase
}

// NewUsecasesInteractor initializes a new usecases interactor
func NewUsecasesInteractor(infrastructure infrastructure.Infrastructure, baseExtension extension.BaseExtension, pinsExtension extension.PINExtension) Interactor {

	profile := NewProfileUseCase(infrastructure, baseExtension)
	login := NewLoginUseCases(infrastructure, profile, baseExtension, pinsExtension)
	roles := NewRoleUseCases(infrastructure, baseExtension)
	pins := NewUserPinUseCase(infrastructure, profile, baseExtension, pinsExtension)
	signup := NewSignUpUseCases(infrastructure, profile, pins, baseExtension)
	surveys := NewSurveyUseCases(infrastructure, baseExtension)
	services := admin.NewService(baseExtension)

	impl := Interactor{
		login,
		profile,
		roles,
		signup,
		surveys,
		pins,
		services,
	}

	return impl
}
