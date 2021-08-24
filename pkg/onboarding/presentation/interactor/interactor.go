// Package interactor represent reusable chunks of code that abstract
// logic from presenters while simplifying your app and making future changes effortless.
package interactor

import (
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/engagement"

	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/crm"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/messaging"
	pubsubmessaging "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/pubsub"
	"github.com/savannahghi/onboarding/pkg/onboarding/usecases"
	"github.com/savannahghi/onboarding/pkg/onboarding/usecases/admin"
	"github.com/savannahghi/onboarding/pkg/onboarding/usecases/ussd"
)

// Interactor represents an assemble of all use cases into a single object that can be instantiated anywhere
type Interactor struct {
	Onboarding usecases.ProfileUseCase
	Signup     usecases.SignUpUseCases
	Supplier   usecases.SupplierUseCases
	Login      usecases.LoginUseCases
	Survey     usecases.SurveyUseCases
	UserPIN    usecases.UserPINUseCases
	Engagement engagement.ServiceEngagement
	Messaging  messaging.ServiceMessaging
	NHIF       usecases.NHIFUseCases
	PubSub     pubsubmessaging.ServicePubSub
	SMS        usecases.SMSUsecase
	AITUSSD    ussd.Usecase
	AdminSrv   admin.Usecase
	CrmExt     crm.ServiceCrm
	Role       usecases.RoleUseCase
}

// NewOnboardingInteractor returns a new onboarding interactor
func NewOnboardingInteractor(
	profile usecases.ProfileUseCase,
	su usecases.SignUpUseCases,
	supplier usecases.SupplierUseCases,
	login usecases.LoginUseCases,
	survey usecases.SurveyUseCases,
	userpin usecases.UserPINUseCases,
	engage engagement.ServiceEngagement,
	mes messaging.ServiceMessaging,
	nhif usecases.NHIFUseCases,
	pubsub pubsubmessaging.ServicePubSub,
	sms usecases.SMSUsecase,
	aitussd ussd.Usecase,
	admin admin.Usecase,
	crmExt crm.ServiceCrm,
	role usecases.RoleUseCase,
) (*Interactor, error) {

	return &Interactor{
		Onboarding: profile,
		Signup:     su,
		Supplier:   supplier,
		Login:      login,
		Survey:     survey,
		UserPIN:    userpin,
		Engagement: engage,
		Messaging:  mes,
		NHIF:       nhif,
		PubSub:     pubsub,
		SMS:        sms,
		AITUSSD:    aitussd,
		AdminSrv:   admin,
		CrmExt:     crmExt,
		Role:       role,
	}, nil
}
