package ussd

import (
	"context"

	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/onboarding/pkg/onboarding/usecases"
	"github.com/savannahghi/profileutils"

	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/extension"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/utils"
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/crm"
	pubsubmessaging "github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/services/pubsub"
	"github.com/savannahghi/onboarding/pkg/onboarding/repository"
)

const (
	// LoginUserState handles workflow required to authenticate/login a user
	LoginUserState = 0
	//HomeMenuState represents inner submenu once user is logged in
	HomeMenuState = 5
	// ChangeUserPINState represents workflows required to set a user PIN
	ChangeUserPINState = 50
	// UserPINResetState represents workflows required to reset a forgotten user PIN
	UserPINResetState = 10
	// EmptyInput is used to load a default menu when user has not supplied any input
	EmptyInput = ""
	// GoBackHomeInput represents the user intention to go back to the main menu
	GoBackHomeInput = "0"
)

//Usecase represent the logic involved in processing USSD requests
type Usecase interface {
	HandleResponseFromUSSDGateway(context context.Context, input *dto.SessionDetails) string
	HandleUserRegistration(ctx context.Context, sessionDetails *domain.USSDLeadDetails, userResponse string) string
	HandleHomeMenu(ctx context.Context, level int, session *domain.USSDLeadDetails, userResponse string) string
	CreateUsddUserProfile(ctx context.Context, phoneNumber string, PIN string, userProfile *dto.UserProfileInput) error
	HandleLogin(ctx context.Context, session *domain.USSDLeadDetails, userResponse string) string
	// session usecases
	GetOrCreateSessionState(ctx context.Context, payload *dto.SessionDetails) (*domain.USSDLeadDetails, error)
	AddAITSessionDetails(ctx context.Context, input *dto.SessionDetails) (*domain.USSDLeadDetails, error)
	StageCRMPayload(ctx context.Context, payload *dto.ContactLeadInput) error
	UpdateSessionLevel(ctx context.Context, level int, sessionID string) error
	UpdateSessionPIN(ctx context.Context, pin string, sessionID string) (*domain.USSDLeadDetails, error)
	// USSD PIN usecases
	HandleChangePIN(ctx context.Context, session *domain.USSDLeadDetails, userResponse string) string
	HandlePINReset(ctx context.Context, session *domain.USSDLeadDetails, userResponse string) string
	SetUSSDUserPin(ctx context.Context, phoneNumber string, PIN string) error
	ChangeUSSDUserPIN(ctx context.Context, phone string, pin string) (bool, error)

	// Onboarding
	GetOrCreatePhoneNumberUser(ctx context.Context, phone string) (*dto.CreatedUserResponse, error)
	CreateUserProfile(ctx context.Context, phoneNumber, uid string) (*profileutils.UserProfile, error)
	CreateEmptyCustomerProfile(ctx context.Context, profileID string) (*profileutils.Customer, error)
	UpdateBioData(ctx context.Context, id string, data profileutils.BioData) error
	GetUserProfileByPrimaryPhoneNumber(ctx context.Context, phoneNumber string, suspend bool) (*profileutils.UserProfile, error)
	RemoveUserByPhoneNumber(ctx context.Context, phone string) error
	SetUserPIN(ctx context.Context, pin string, profileID string) (bool, error)
	//USSD Metrics Event
	SaveUSSDEvent(ctx context.Context, input *dto.USSDEvent) (*dto.USSDEvent, error)
}

//Impl represents usecase implementation
type Impl struct {
	baseExt              extension.BaseExtension
	onboardingRepository repository.OnboardingRepository
	profile              usecases.ProfileUseCase
	pinUsecase           usecases.UserPINUseCases
	signUp               usecases.SignUpUseCases
	pinExt               extension.PINExtension
	pubsub               pubsubmessaging.ServicePubSub
	crm                  crm.ServiceCrm
}

//NewUssdUsecases returns a new USSD usecase
func NewUssdUsecases(
	repository repository.OnboardingRepository,
	ext extension.BaseExtension,
	profileUsecase usecases.ProfileUseCase,
	pinUsecase usecases.UserPINUseCases,
	signUp usecases.SignUpUseCases,
	pinExt extension.PINExtension,
	pubsub pubsubmessaging.ServicePubSub,
	crm crm.ServiceCrm,
) Usecase {
	return &Impl{
		baseExt:              ext,
		onboardingRepository: repository,
		profile:              profileUsecase,
		pinUsecase:           pinUsecase,
		signUp:               signUp,
		pinExt:               pinExt,
		pubsub:               pubsub,
		crm:                  crm,
	}
}

//HandleResponseFromUSSDGateway receives and processes the USSD response from the USSD gateway
func (u *Impl) HandleResponseFromUSSDGateway(ctx context.Context, payload *dto.SessionDetails) string {
	ctx, span := tracer.Start(ctx, "HandleResponseFromUSSDGateway")
	defer span.End()

	sessionDetails, err := u.GetOrCreateSessionState(ctx, payload)
	if err != nil {
		utils.RecordSpanError(span, err)
		return "END Something went wrong. Please try again."
	}

	userResponse := utils.GetUserResponse(payload.Text)

	exists, err := u.profile.CheckPhoneExists(ctx, *payload.PhoneNumber)
	if err != nil {
		utils.RecordSpanError(span, err)
		return "END Something went wrong. Please try again."
	}

	if !exists {
		return u.HandleUserRegistration(ctx, sessionDetails, userResponse)
	}

	switch {
	case sessionDetails.Level == LoginUserState:
		return u.HandleLogin(ctx, sessionDetails, userResponse)

	case sessionDetails.Level == HomeMenuState:
		return u.HandleHomeMenu(ctx, HomeMenuState, sessionDetails, userResponse)

	case sessionDetails.Level >= ChangeUserPINState:
		return u.HandleChangePIN(ctx, sessionDetails, userResponse)

	case sessionDetails.Level >= UserPINResetState:
		return u.HandlePINReset(ctx, sessionDetails, userResponse)

	default:
		return "END Something went wrong. Please try again."
	}

}
