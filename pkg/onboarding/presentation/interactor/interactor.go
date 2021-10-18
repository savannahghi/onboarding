// Package interactor represent reusable chunks of code that abstract
// logic from presenters while simplifying your app and making future changes effortless.
package interactor

import (
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/chargemaster"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/edi"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/engagement"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/messaging"
	pubsubmessaging "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/pubsub"
	"github.com/savannahghi/onboarding/pkg/onboarding/usecases"
	"github.com/savannahghi/onboarding/pkg/onboarding/usecases/admin"
	"github.com/savannahghi/onboarding/pkg/onboarding/usecases/ussd"
	erp "gitlab.slade360emr.com/go/commontools/accounting/pkg/usecases"
)

// Interactor represents an assemble of all use cases into a single object that can be instantiated anywhere
type Interactor struct {
	Onboarding   usecases.ProfileUseCase
	Signup       usecases.SignUpUseCases
	Supplier     usecases.SupplierUseCases
	Login        usecases.LoginUseCases
	Survey       usecases.SurveyUseCases
	UserPIN      usecases.UserPINUseCases
	ERP          erp.AccountingUsecase
	ChargeMaster chargemaster.ServiceChargeMaster
	Engagement   engagement.ServiceEngagement
	Messaging    messaging.ServiceMessaging
	NHIF         usecases.NHIFUseCases
	PubSub       pubsubmessaging.ServicePubSub
	SMS          usecases.SMSUsecase
	AITUSSD      ussd.Usecase
	Agent        usecases.AgentUseCase
	Admin        usecases.AdminUseCase
	EDI          edi.ServiceEdi
	AdminSrv     admin.Usecase
	Role         usecases.RoleUseCase
}

// NewOnboardingInteractor returns a new onboarding interactor
func NewOnboardingInteractor(
	profile usecases.ProfileUseCase,
	su usecases.SignUpUseCases,
	supplier usecases.SupplierUseCases,
	login usecases.LoginUseCases,
	survey usecases.SurveyUseCases,
	userpin usecases.UserPINUseCases,
	erp erp.AccountingUsecase,
	chrg chargemaster.ServiceChargeMaster,
	engage engagement.ServiceEngagement,
	mes messaging.ServiceMessaging,
	nhif usecases.NHIFUseCases,
	pubsub pubsubmessaging.ServicePubSub,
	sms usecases.SMSUsecase,
	aitussd ussd.Usecase,
	agt usecases.AgentUseCase,
	adm usecases.AdminUseCase,
	edi edi.ServiceEdi,
	admin admin.Usecase,
	role usecases.RoleUseCase,
) (*Interactor, error) {

	return &Interactor{
		Onboarding:   profile,
		Signup:       su,
		Supplier:     supplier,
		Login:        login,
		Survey:       survey,
		UserPIN:      userpin,
		ERP:          erp,
		ChargeMaster: chrg,
		Engagement:   engage,
		Messaging:    mes,
		NHIF:         nhif,
		PubSub:       pubsub,
		SMS:          sms,
		AITUSSD:      aitussd,
		Agent:        agt,
		Admin:        adm,
		EDI:          edi,
		AdminSrv:     admin,
		Role:         role,
	}, nil
}
