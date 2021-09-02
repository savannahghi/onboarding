// Package interactor represent reusable chunks of code that abstract
// logic from presenters while simplifying your app and making future changes effortless.
package interactor

import (
	"github.com/savannahghi/onboarding/pkg/onboarding/usecases"
	"github.com/savannahghi/onboarding/pkg/onboarding/usecases/admin"
)

// Interactor represents an assemble of all use cases into a single object that can be instantiated anywhere
type Interactor struct {
	Onboarding usecases.ProfileUseCase
	Signup     usecases.SignUpUseCases
	Login      usecases.LoginUseCases
	Survey     usecases.SurveyUseCases
	UserPIN    usecases.UserPINUseCases
	AdminSrv   admin.Usecase
	Role       usecases.RoleUseCase
}

// NewOnboardingInteractor returns a new onboarding interactor
func NewOnboardingInteractor(
	profile usecases.ProfileUseCase,
	su usecases.SignUpUseCases,
	login usecases.LoginUseCases,
	survey usecases.SurveyUseCases,
	userpin usecases.UserPINUseCases,
	admin admin.Usecase,
	role usecases.RoleUseCase,
) (*Interactor, error) {

	return &Interactor{
		Onboarding: profile,
		Signup:     su,
		Login:      login,
		Survey:     survey,
		UserPIN:    userpin,
		AdminSrv:   admin,
		Role:       role,
	}, nil
}
