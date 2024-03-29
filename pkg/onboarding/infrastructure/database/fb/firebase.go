package fb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"firebase.google.com/go/auth"
	"github.com/google/uuid"

	"github.com/savannahghi/converterandformatter"
	"github.com/savannahghi/enumutils"
	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/firebasetools"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/exceptions"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/utils"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/profileutils"
	"github.com/savannahghi/pubsubtools"
	"github.com/savannahghi/scalarutils"
	"github.com/savannahghi/serverutils"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"golang.org/x/time/rate"
)

// Package that generates trace information
var tracer = otel.Tracer(
	"github.com/savannahghi/onboarding/pkg/onboarding/infrastructure/database/fb",
)

const (
	userProfileCollectionName            = "user_profiles"
	pinsCollectionName                   = "pins"
	surveyCollectionName                 = "post_visit_survey"
	profileNudgesCollectionName          = "profile_nudges"
	experimentParticipantCollectionName  = "experiment_participants"
	communicationsSettingsCollectionName = "communications_settings"
	firebaseExchangeRefreshTokenURL      = "https://securetoken.googleapis.com/v1/token?key="
	rolesRevocationCollectionName        = "role_revocations"
	rolesCollectionName                  = "user_roles"
)

// Repository accesses and updates an item that is stored on Firebase
type Repository struct {
	FirestoreClient FirestoreClientExtension
	FirebaseClient  FirebaseClientExtension
}

// NewFirebaseRepository initializes a Firebase repository
func NewFirebaseRepository(
	firestoreClient FirestoreClientExtension,
	firebaseClient FirebaseClientExtension,
) *Repository {
	return &Repository{
		FirestoreClient: firestoreClient,
		FirebaseClient:  firebaseClient,
	}
}

// GetUserProfileCollectionName ...
func (fr Repository) GetUserProfileCollectionName() string {
	suffixed := firebasetools.SuffixCollection(userProfileCollectionName)
	return suffixed
}

// GetSurveyCollectionName returns a well suffixed PINs collection name
func (fr Repository) GetSurveyCollectionName() string {
	suffixed := firebasetools.SuffixCollection(surveyCollectionName)
	return suffixed
}

// GetPINsCollectionName returns a well suffixed PINs collection name
func (fr Repository) GetPINsCollectionName() string {
	suffixed := firebasetools.SuffixCollection(pinsCollectionName)
	return suffixed
}

// GetProfileNudgesCollectionName return the storage location of profile nudges
func (fr Repository) GetProfileNudgesCollectionName() string {
	suffixed := firebasetools.SuffixCollection(profileNudgesCollectionName)
	return suffixed
}

// GetExperimentParticipantCollectionName fetches the collection where experiment participant will be saved
func (fr *Repository) GetExperimentParticipantCollectionName() string {
	suffixed := firebasetools.SuffixCollection(experimentParticipantCollectionName)
	return suffixed
}

// GetCommunicationsSettingsCollectionName ...
func (fr Repository) GetCommunicationsSettingsCollectionName() string {
	suffixed := firebasetools.SuffixCollection(communicationsSettingsCollectionName)
	return suffixed
}

// GetRolesCollectionName ...
func (fr Repository) GetRolesCollectionName() string {
	suffixed := firebasetools.SuffixCollection(rolesCollectionName)
	return suffixed
}

// GetRolesRevocationCollectionName ...
func (fr Repository) GetRolesRevocationCollectionName() string {
	suffixed := firebasetools.SuffixCollection(rolesRevocationCollectionName)
	return suffixed
}

// GetUserProfileByUID retrieves the user profile by UID
func (fr *Repository) GetUserProfileByUID(
	ctx context.Context,
	uid string,
	suspended bool,
) (*profileutils.UserProfile, error) {
	ctx, span := tracer.Start(ctx, "GetUserProfileByUID")
	defer span.End()

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "verifiedUIDS",
		Value:          uid,
		Operator:       "array-contains",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(err)
	}
	if len(docs) == 0 {
		err = exceptions.ProfileNotFoundError(fmt.Errorf("user profile not found"))

		utils.RecordSpanError(span, err)
		return nil, err
	}

	if len(docs) > 1 && serverutils.IsDebug() {
		log.Printf("user with uids %s has > 1 profile (they have %d)",
			uid,
			len(docs),
		)
	}

	dsnap := docs[0]
	userProfile := &profileutils.UserProfile{}
	err = dsnap.DataTo(userProfile)
	if err != nil {
		utils.RecordSpanError(span, err)
		err = fmt.Errorf("unable to read user profile")
		return nil, exceptions.InternalServerError(err)
	}

	if !suspended {
		// never return a suspended user profile
		if userProfile.Suspended {
			return nil, exceptions.ProfileSuspendFoundError()
		}
	}

	return userProfile, nil
}

//GetUserProfileByPhoneOrEmail retrieves user profile by email adddress
func (fr *Repository) GetUserProfileByPhoneOrEmail(ctx context.Context, payload *dto.RetrieveUserProfileInput) (*profileutils.UserProfile, error) {
	ctx, span := tracer.Start(ctx, "GetUserProfileByPhoneOrEmail")
	defer span.End()

	if payload.PhoneNumber == nil {
		query := &GetAllQuery{
			CollectionName: fr.GetUserProfileCollectionName(),
			FieldName:      "primaryEmailAddress",
			Value:          payload.Email,
			Operator:       "==",
		}

		docs, err := fr.FirestoreClient.GetAll(ctx, query)
		if err != nil {
			utils.RecordSpanError(span, err)
			return nil, exceptions.InternalServerError(err)
		}

		if len(docs) == 0 {
			query := &GetAllQuery{
				CollectionName: fr.GetUserProfileCollectionName(),
				FieldName:      "secondaryEmailAddresses",
				Value:          payload.Email,
				Operator:       "array-contains",
			}

			docs, err := fr.FirestoreClient.GetAll(ctx, query)
			if err != nil {
				utils.RecordSpanError(span, err)
				return nil, exceptions.InternalServerError(err)
			}

			if len(docs) == 0 {
				err = exceptions.ProfileNotFoundError(err)

				utils.RecordSpanError(span, err)
				return nil, err
			}

			dsnap := docs[0]
			userProfile := &profileutils.UserProfile{}
			err = dsnap.DataTo(userProfile)
			if err != nil {
				utils.RecordSpanError(span, err)
				err = fmt.Errorf("unable to read user profile")
				return nil, exceptions.InternalServerError(err)
			}

			return userProfile, nil
		}

		dsnap := docs[0]
		userProfile := &profileutils.UserProfile{}
		err = dsnap.DataTo(userProfile)
		if err != nil {
			utils.RecordSpanError(span, err)
			err = fmt.Errorf("unable to read user profile")
			return nil, exceptions.InternalServerError(err)
		}

		return userProfile, nil
	}

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "primaryPhone",
		Value:          payload.PhoneNumber,
		Operator:       "==",
	}

	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(err)
	}

	if len(docs) == 0 {
		query := &GetAllQuery{
			CollectionName: fr.GetUserProfileCollectionName(),
			FieldName:      "secondaryPhoneNumbers",
			Value:          payload.PhoneNumber,
			Operator:       "array-contains",
		}

		docs, err := fr.FirestoreClient.GetAll(ctx, query)
		if err != nil {
			utils.RecordSpanError(span, err)
			return nil, exceptions.InternalServerError(err)
		}

		if len(docs) == 0 {
			err = exceptions.ProfileNotFoundError(err)

			utils.RecordSpanError(span, err)
			return nil, err
		}

		dsnap := docs[0]
		userProfile := &profileutils.UserProfile{}
		err = dsnap.DataTo(userProfile)
		if err != nil {
			utils.RecordSpanError(span, err)
			err = fmt.Errorf("unable to read user profile")
			return nil, exceptions.InternalServerError(err)
		}

		return userProfile, nil
	}

	dsnap := docs[0]
	userProfile := &profileutils.UserProfile{}
	err = dsnap.DataTo(userProfile)
	if err != nil {
		utils.RecordSpanError(span, err)
		err = fmt.Errorf("unable to read user profile")
		return nil, exceptions.InternalServerError(err)
	}

	return userProfile, nil
}

// UpdateUserProfileEmail updates user profile's email
func (fr *Repository) UpdateUserProfileEmail(
	ctx context.Context,
	phone string,
	email string,
) error {
	ctx, span := tracer.Start(ctx, "UpdateUserProfileEmail")
	defer span.End()

	payload := &dto.RetrieveUserProfileInput{
		PhoneNumber: &phone,
	}

	profile, err := fr.GetUserProfileByPhoneOrEmail(ctx, payload)
	if err != nil {
		utils.RecordSpanError(span, err)
		// this is a wrapped error. No need to wrap it again
		return err
	}
	profile.PrimaryEmailAddress = &email

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "primaryPhone",
		Value:          profile.PrimaryPhone,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to parse user profile as firebase snapshot: %v", err),
		)
	}

	if len(docs) == 0 {
		return exceptions.InternalServerError(fmt.Errorf("user profile not found"))
	}

	updateCommand := &UpdateCommand{
		CollectionName: fr.GetUserProfileCollectionName(),
		ID:             docs[0].Ref.ID,
		Data:           profile,
	}
	err = fr.FirestoreClient.Update(ctx, updateCommand)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to update user profile primary phone number: %v", err),
		)
	}

	return nil
}

// GetUserProfileByID retrieves a user profile by ID
func (fr *Repository) GetUserProfileByID(
	ctx context.Context,
	id string,
	suspended bool,
) (*profileutils.UserProfile, error) {
	ctx, span := tracer.Start(ctx, "GetUserProfileByID")
	defer span.End()

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "id",
		Value:          id,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(err)
	}
	if len(docs) > 1 && serverutils.IsDebug() {
		log.Printf("> 1 profile with id %s (count: %d)", id, len(docs))
	}

	if len(docs) == 0 {
		return nil, exceptions.ProfileNotFoundError(fmt.Errorf("user profile not found"))
	}
	dsnap := docs[0]
	userProfile := &profileutils.UserProfile{}
	err = dsnap.DataTo(userProfile)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(
			fmt.Errorf("unable to read user profile: %w", err),
		)
	}

	if !suspended {
		// never return a suspended user profile
		if userProfile.Suspended {
			return nil, exceptions.ProfileSuspendFoundError()
		}
	}
	return userProfile, nil
}

func (fr *Repository) fetchUserRandomName(ctx context.Context) *string {
	n := utils.GetRandomName()
	if v, err := fr.CheckIfUsernameExists(ctx, *n); v && (err == nil) {
		return fr.fetchUserRandomName(ctx)
	}
	return n
}

// CreateUserProfile creates a user profile of using the provided phone number and uid
func (fr *Repository) CreateUserProfile(
	ctx context.Context,
	phoneNumber, uid string,
) (*profileutils.UserProfile, error) {
	ctx, span := tracer.Start(ctx, "CreateUserProfile")
	defer span.End()

	v, err := fr.CheckIfPhoneNumberExists(ctx, phoneNumber)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(
			fmt.Errorf("failed to check the phone number: %v", err),
		)
	}

	if v {
		// this phone is number is associated with another user profile, hence can not create an profile with the same phone number
		return nil, exceptions.CheckPhoneNumberExistError()
	}

	profileID := uuid.New().String()
	pr := &profileutils.UserProfile{
		ID:           profileID,
		UserName:     fr.fetchUserRandomName(ctx),
		PrimaryPhone: &phoneNumber,
		VerifiedIdentifiers: []profileutils.VerifiedIdentifier{{
			UID:           uid,
			LoginProvider: profileutils.LoginProviderTypePhone,
			Timestamp:     time.Now().In(pubsubtools.TimeLocation),
		}},
		VerifiedUIDS:  []string{uid},
		TermsAccepted: true,
		Suspended:     false,
	}

	command := &CreateCommand{
		CollectionName: fr.GetUserProfileCollectionName(),
		Data:           pr,
	}
	docRef, err := fr.FirestoreClient.Create(ctx, command)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(
			fmt.Errorf("unable to create new user profile: %w", err),
		)
	}
	query := &GetSingleQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		Value:          docRef.ID,
	}
	dsnap, err := fr.FirestoreClient.Get(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(
			fmt.Errorf("unable to retrieve newly created user profile: %w", err),
		)
	}
	// return the newly created user profile
	userProfile := &profileutils.UserProfile{}
	err = dsnap.DataTo(userProfile)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(
			fmt.Errorf("unable to read user profile: %w", err),
		)
	}
	return userProfile, nil

}

// CreateDetailedUserProfile creates a new user profile that is pre-filled using the provided phone number
func (fr *Repository) CreateDetailedUserProfile(
	ctx context.Context,
	phoneNumber string,
	profile profileutils.UserProfile,
) (*profileutils.UserProfile, error) {
	ctx, span := tracer.Start(ctx, "CreateDetailedUserProfile")
	defer span.End()

	exists, err := fr.CheckIfPhoneNumberExists(ctx, phoneNumber)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(
			fmt.Errorf("failed to check the phone number: %v", err),
		)
	}

	if exists {
		// this phone is number is associated with another user profile, hence can not create an profile with the same phone number
		err = exceptions.CheckPhoneNumberExistError()
		utils.RecordSpanError(span, err)
		return nil, err
	}

	// create user via their phone number on firebase
	user, err := fr.GetOrCreatePhoneNumberUser(ctx, phoneNumber)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}

	phoneIdentifier := profileutils.VerifiedIdentifier{
		UID:           user.UID,
		LoginProvider: profileutils.LoginProviderTypePhone,
		Timestamp:     time.Now().In(pubsubtools.TimeLocation),
	}

	profile.VerifiedIdentifiers = append(profile.VerifiedIdentifiers, phoneIdentifier)
	profile.VerifiedUIDS = append(profile.VerifiedUIDS, user.UID)

	profileID := uuid.New().String()
	profile.ID = profileID
	profile.PrimaryPhone = &phoneNumber
	profile.UserName = fr.fetchUserRandomName(ctx)
	profile.TermsAccepted = true
	profile.Suspended = false

	command := &CreateCommand{
		CollectionName: fr.GetUserProfileCollectionName(),
		Data:           profile,
	}

	_, err = fr.FirestoreClient.Create(ctx, command)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(
			fmt.Errorf("unable to create new user profile: %w", err),
		)
	}

	return &profile, nil
}

//GetUserProfileByPrimaryPhoneNumber fetches a user profile by primary phone number
func (fr *Repository) GetUserProfileByPrimaryPhoneNumber(
	ctx context.Context,
	phoneNumber string,
	suspended bool,
) (*profileutils.UserProfile, error) {
	ctx, span := tracer.Start(ctx, "GetUserProfileByPrimaryPhoneNumber")
	defer span.End()

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "primaryPhone",
		Value:          phoneNumber,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(err)
	}
	if len(docs) == 0 {
		return nil, exceptions.ProfileNotFoundError(fmt.Errorf("user profile not found"))
	}
	dsnap := docs[0]
	profile := &profileutils.UserProfile{}
	err = dsnap.DataTo(profile)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(
			fmt.Errorf("unable to read user profile: %w", err),
		)
	}

	if !suspended {
		// never return a suspended user profile
		if profile.Suspended {
			return nil, exceptions.ProfileSuspendFoundError()
		}
	}
	return profile, nil
}

// GetUserProfileByPhoneNumber fetches a user profile by phone number. This method traverses both PRIMARY PHONE numbers
// and SECONDARY PHONE numbers.
func (fr *Repository) GetUserProfileByPhoneNumber(
	ctx context.Context,
	phoneNumber string,
	suspended bool,
) (*profileutils.UserProfile, error) {
	ctx, span := tracer.Start(ctx, "GetUserProfileByPhoneNumber")
	defer span.End()

	// check first primary phone numbers
	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "primaryPhone",
		Value:          phoneNumber,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(err)
	}
	if len(docs) == 1 {
		dsnap := docs[0]
		pr := &profileutils.UserProfile{}
		if err := dsnap.DataTo(pr); err != nil {
			return nil, exceptions.InternalServerError(
				fmt.Errorf("unable to read user profile: %w", err),
			)
		}
		return pr, nil
	}

	// then check in secondary phone numbers
	query1 := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "secondaryPhoneNumbers",
		Value:          phoneNumber,
		Operator:       "array-contains",
	}
	docs1, err := fr.FirestoreClient.GetAll(ctx, query1)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(err)
	}

	if len(docs1) == 1 {
		dsnap := docs1[0]
		pr := &profileutils.UserProfile{}
		if err := dsnap.DataTo(pr); err != nil {
			return nil, exceptions.InternalServerError(
				fmt.Errorf("unable to read user profile: %w", err),
			)
		}

		if !suspended {
			// never return a suspended user profile
			if pr.Suspended {
				return nil, exceptions.ProfileSuspendFoundError()
			}
		}

		return pr, nil
	}

	return nil, exceptions.ProfileNotFoundError(fmt.Errorf("user profile not found"))

}

// CheckIfPhoneNumberExists checks both PRIMARY PHONE NUMBERs and SECONDARY PHONE NUMBERs for the
// existence of the argument phoneNumber.
func (fr *Repository) CheckIfPhoneNumberExists(
	ctx context.Context,
	phoneNumber string,
) (bool, error) {
	ctx, span := tracer.Start(ctx, "CheckIfPhoneNumberExists")
	defer span.End()

	// check first primary phone numbers
	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "primaryPhone",
		Value:          phoneNumber,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.InternalServerError(err)
	}

	if len(docs) > 0 {
		return true, nil
	}

	// then check in secondary phone numbers
	query1 := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "secondaryPhoneNumbers",
		Value:          phoneNumber,
		Operator:       "array-contains",
	}
	docs1, err := fr.FirestoreClient.GetAll(ctx, query1)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.InternalServerError(err)
	}
	if len(docs1) > 0 {
		return true, nil
	}

	return false, nil
}

// CheckIfEmailExists checks in both PRIMARY EMAIL and SECONDARY EMAIL for the
// existence of the argument email
func (fr *Repository) CheckIfEmailExists(ctx context.Context, email string) (bool, error) {
	ctx, span := tracer.Start(ctx, "CheckIfEmailExists")
	defer span.End()

	// check first primary email
	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "primaryEmailAddress",
		Value:          email,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.InternalServerError(err)
	}
	if len(docs) == 1 {
		return true, nil
	}

	// then check in secondary email
	query1 := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "secondaryEmailAddresses",
		Value:          email,
		Operator:       "array-contains",
	}
	docs1, err := fr.FirestoreClient.GetAll(ctx, query1)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, err
	}
	if len(docs1) == 1 {
		return true, nil
	}
	return false, nil
}

// CheckIfUsernameExists checks if the provided username exists. If true, it means its has already been associated with
// another user
func (fr *Repository) CheckIfUsernameExists(ctx context.Context, userName string) (bool, error) {
	ctx, span := tracer.Start(ctx, "CheckIfUsernameExists")
	defer span.End()

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "userName",
		Value:          userName,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.InternalServerError(err)
	}
	if len(docs) == 1 {
		return true, nil
	}

	return false, nil
}

// GetPINByProfileID gets a user's PIN by their profile ID
func (fr *Repository) GetPINByProfileID(
	ctx context.Context,
	profileID string,
) (*domain.PIN, error) {
	ctx, span := tracer.Start(ctx, "GetPINByProfileID")
	defer span.End()

	query := &GetAllQuery{
		CollectionName: fr.GetPINsCollectionName(),
		FieldName:      "profileID",
		Value:          profileID,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(err)
	}
	// this should never run. If it does, it means we are doing something wrong.
	if len(docs) > 1 && serverutils.IsDebug() {
		log.Printf("> 1 PINs with profile ID %s (count: %d)", profileID, len(docs))
	}

	if len(docs) == 0 {
		return nil, exceptions.PinNotFoundError(fmt.Errorf("failed to get a user pin"))
	}

	dsnap := docs[0]
	PIN := &domain.PIN{}
	err = dsnap.DataTo(PIN)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}

	return PIN, nil
}

// GenerateAuthCredentialsForAnonymousUser generates auth credentials for the anonymous user. This method is here since we don't
// want to delegate sign-in of anonymous users to the frontend. This is an effort the over reliance on firebase and lettin us
// handle all the heavy lifting
func (fr *Repository) GenerateAuthCredentialsForAnonymousUser(
	ctx context.Context,
) (*profileutils.AuthCredentialResponse, error) {
	ctx, span := tracer.Start(ctx, "GenerateAuthCredentialsForAnonymousUser")
	defer span.End()

	// todo(dexter) : move anonymousPhoneNumber to base. AnonymousPhoneNumber should NEVER NEVER have a user profile
	anonymousPhoneNumber := "+254700000000"

	u, err := fr.GetOrCreatePhoneNumberUser(ctx, anonymousPhoneNumber)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(err)
	}

	customToken, err := firebasetools.CreateFirebaseCustomToken(ctx, u.UID)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.CustomTokenError(err)
	}
	userTokens, err := firebasetools.AuthenticateCustomFirebaseToken(customToken)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.AuthenticateTokenError(err)
	}

	return &profileutils.AuthCredentialResponse{
		CustomToken:  &customToken,
		IDToken:      &userTokens.IDToken,
		ExpiresIn:    userTokens.ExpiresIn,
		RefreshToken: userTokens.RefreshToken,
		UID:          u.UID,
		IsAnonymous:  true,
		IsAdmin:      false,
	}, nil
}

// GenerateAuthCredentials gets a Firebase user by phone and creates their tokens
func (fr *Repository) GenerateAuthCredentials(
	ctx context.Context,
	phone string,
	profile *profileutils.UserProfile,
) (*profileutils.AuthCredentialResponse, error) {
	ctx, span := tracer.Start(ctx, "GenerateAuthCredentials")
	defer span.End()

	resp, err := fr.GetOrCreatePhoneNumberUser(ctx, phone)
	if err != nil {
		utils.RecordSpanError(span, err)
		if auth.IsUserNotFound(err) {
			return nil, exceptions.UserNotFoundError(err)
		}
		return nil, exceptions.UserNotFoundError(err)
	}

	customToken, err := firebasetools.CreateFirebaseCustomToken(ctx, resp.UID)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.CustomTokenError(err)
	}
	userTokens, err := firebasetools.AuthenticateCustomFirebaseToken(customToken)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.AuthenticateTokenError(err)
	}

	if err := fr.UpdateVerifiedIdentifiers(ctx, profile.ID, []profileutils.VerifiedIdentifier{{
		UID:           resp.UID,
		LoginProvider: profileutils.LoginProviderTypePhone,
		Timestamp:     time.Now().In(pubsubtools.TimeLocation),
	}}); err != nil {
		return nil, exceptions.UpdateProfileError(err)
	}

	if err := fr.UpdateVerifiedUIDS(ctx, profile.ID, []string{resp.UID}); err != nil {
		return nil, exceptions.UpdateProfileError(err)
	}

	canExperiment, err := fr.CheckIfExperimentParticipant(ctx, profile.ID)
	if err != nil {
		utils.RecordSpanError(span, err)
		// this is a wrapped error. No need to wrap it again
		return nil, err
	}

	return &profileutils.AuthCredentialResponse{
		CustomToken:   &customToken,
		IDToken:       &userTokens.IDToken,
		ExpiresIn:     userTokens.ExpiresIn,
		RefreshToken:  userTokens.RefreshToken,
		UID:           resp.UID,
		IsAnonymous:   false,
		IsAdmin:       fr.CheckIfAdmin(profile),
		CanExperiment: canExperiment,
	}, nil
}

// CheckIfAdmin checks if a user has admin permissions
func (fr *Repository) CheckIfAdmin(profile *profileutils.UserProfile) bool {
	if len(profile.Permissions) == 0 {
		return false
	}
	exists := false
	for _, p := range profile.Permissions {
		if p == profileutils.PermissionTypeSuperAdmin || p == profileutils.PermissionTypeAdmin {
			exists = true
			break
		}
	}
	return exists
}

// UpdateUserName updates the username of a profile that matches the id
// this method should be called after asserting the username is unique and not associated with another userProfile
func (fr *Repository) UpdateUserName(ctx context.Context, id string, userName string) error {
	ctx, span := tracer.Start(ctx, "UpdateUserName")
	defer span.End()

	v, err := fr.CheckIfUsernameExists(ctx, userName)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(err)
	}
	if v {
		return exceptions.InternalServerError(fmt.Errorf("%v", exceptions.UsernameInUseErrMsg))
	}
	profile, err := fr.GetUserProfileByID(ctx, id, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		// this is a wrapped error. No need to wrap it again
		return err
	}
	profile.UserName = &userName
	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "id",
		Value:          profile.ID,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to parse user profile as firebase snapshot: %v", err),
		)
	}

	if len(docs) == 0 {
		return exceptions.InternalServerError(fmt.Errorf("user profile not found"))
	}
	updateCommand := &UpdateCommand{
		CollectionName: fr.GetUserProfileCollectionName(),
		ID:             docs[0].Ref.ID,
		Data:           profile,
	}
	err = fr.FirestoreClient.Update(ctx, updateCommand)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to update user profile primary phone number: %v", err),
		)
	}

	return nil
}

// UpdatePrimaryPhoneNumber append a new primary phone number to the user profile
// this method should be called after asserting the phone number is unique and not associated with another userProfile
func (fr *Repository) UpdatePrimaryPhoneNumber(
	ctx context.Context,
	id string,
	phoneNumber string,
) error {
	ctx, span := tracer.Start(ctx, "UpdatePrimaryPhoneNumber")
	defer span.End()

	profile, err := fr.GetUserProfileByID(ctx, id, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		// this is a wrapped error. No need to wrap it again
		return err
	}
	profile.PrimaryPhone = &phoneNumber

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "id",
		Value:          profile.ID,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to parse user profile as firebase snapshot: %v", err),
		)
	}

	if len(docs) == 0 {
		return exceptions.InternalServerError(fmt.Errorf("user profile not found"))
	}

	updateCommand := &UpdateCommand{
		CollectionName: fr.GetUserProfileCollectionName(),
		ID:             docs[0].Ref.ID,
		Data:           profile,
	}
	err = fr.FirestoreClient.Update(ctx, updateCommand)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to update user profile primary phone number: %v", err),
		)
	}

	return nil
}

// UpdateUserRoleIDs updates the roles for a user
func (fr Repository) UpdateUserRoleIDs(ctx context.Context, id string, roleIDs []string) error {
	ctx, span := tracer.Start(ctx, "UpdateUserRoleIDs")
	defer span.End()

	profile, err := fr.GetUserProfileByID(ctx, id, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		return err
	}

	// Add the roles
	profile.Roles = roleIDs

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "id",
		Value:          profile.ID,
		Operator:       "==",
	}

	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(err)
	}

	if len(docs) == 0 {
		return exceptions.InternalServerError(fmt.Errorf("user profile not found"))
	}

	updateCommand := &UpdateCommand{
		CollectionName: fr.GetUserProfileCollectionName(),
		ID:             docs[0].Ref.ID,
		Data:           profile,
	}

	err = fr.FirestoreClient.Update(ctx, updateCommand)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to update user profile primary email address: %v", err),
		)
	}

	return nil
}

// UpdatePrimaryEmailAddress the primary email addresses of the profile that matches the id
// this method should be called after asserting the emailAddress is unique and not associated with another userProfile
func (fr *Repository) UpdatePrimaryEmailAddress(
	ctx context.Context,
	id string,
	emailAddress string,
) error {
	ctx, span := tracer.Start(ctx, "UpdatePrimaryEmailAddress")
	defer span.End()

	profile, err := fr.GetUserProfileByID(ctx, id, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		// this is a wrapped error. No need to wrap it again
		return err
	}
	profile.PrimaryEmailAddress = &emailAddress

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "id",
		Value:          profile.ID,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to parse user profile as firebase snapshot: %v", err),
		)
	}

	if len(docs) == 0 {
		return exceptions.InternalServerError(fmt.Errorf("user profile not found"))
	}
	updateCommand := &UpdateCommand{
		CollectionName: fr.GetUserProfileCollectionName(),
		ID:             docs[0].Ref.ID,
		Data:           profile,
	}
	err = fr.FirestoreClient.Update(ctx, updateCommand)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to update user profile primary email address: %v", err),
		)
	}

	return nil
}

// UpdateSecondaryPhoneNumbers updates the secondary phone numbers of the profile that matches the id
// this method should be called after asserting the phone numbers are unique and not associated with another userProfile
func (fr *Repository) UpdateSecondaryPhoneNumbers(
	ctx context.Context,
	id string,
	phoneNumbers []string,
) error {
	ctx, span := tracer.Start(ctx, "UpdateSecondaryPhoneNumbers")
	defer span.End()

	profile, err := fr.GetUserProfileByID(ctx, id, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		// this is a wrapped error. No need to wrap it again
		return err
	}

	// Check if the former primary phone exists in the phoneNumber list
	index, exist := utils.FindItem(profile.SecondaryPhoneNumbers, *profile.PrimaryPhone)
	if exist {
		// Remove the former secondary phone from the list since it's now primary
		profile.SecondaryPhoneNumbers = append(
			profile.SecondaryPhoneNumbers[:index],
			profile.SecondaryPhoneNumbers[index+1:]...,
		)
	}

	for _, phone := range phoneNumbers {
		index, exist := utils.FindItem(profile.SecondaryPhoneNumbers, phone)
		if exist {
			profile.SecondaryPhoneNumbers = append(
				profile.SecondaryPhoneNumbers[:index],
				profile.SecondaryPhoneNumbers[index+1:]...)
		}
	}

	profile.SecondaryPhoneNumbers = append(profile.SecondaryPhoneNumbers, phoneNumbers...)

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "id",
		Value:          profile.ID,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to parse user profile as firebase snapshot: %v", err),
		)
	}

	if len(docs) == 0 {
		return exceptions.InternalServerError(fmt.Errorf("user profile not found"))
	}
	updateCommand := &UpdateCommand{
		CollectionName: fr.GetUserProfileCollectionName(),
		ID:             docs[0].Ref.ID,
		Data:           profile,
	}
	err = fr.FirestoreClient.Update(ctx, updateCommand)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to update user profile secondary phone numbers: %v", err),
		)
	}

	return nil
}

// UpdateSecondaryEmailAddresses the secondary email addresses of the profile that matches the id
// this method should be called after asserting the emailAddresses  as unique and not associated with another userProfile
func (fr *Repository) UpdateSecondaryEmailAddresses(
	ctx context.Context,
	id string,
	uniqueEmailAddresses []string,
) error {
	ctx, span := tracer.Start(ctx, "UpdateSecondaryEmailAddresses")
	defer span.End()

	profile, err := fr.GetUserProfileByID(ctx, id, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		// this is a wrapped error. No need to wrap it again
		return err
	}

	// check if former primary email still exists in the
	// secondary emails list
	if profile.PrimaryEmailAddress != nil {
		index, exist := utils.FindItem(
			profile.SecondaryEmailAddresses,
			*profile.PrimaryEmailAddress,
		)
		if exist {
			// remove the former secondary email from the list
			profile.SecondaryEmailAddresses = append(
				profile.SecondaryEmailAddresses[:index],
				profile.SecondaryEmailAddresses[index+1:]...,
			)
		}
	}

	// Check to see whether the new emails exist in list of secondary emails
	for _, email := range uniqueEmailAddresses {
		index, exist := utils.FindItem(profile.SecondaryEmailAddresses, email)
		if exist {
			profile.SecondaryEmailAddresses = append(
				profile.SecondaryEmailAddresses[:index],
				profile.SecondaryEmailAddresses[index+1:]...)
		}
	}

	profile.SecondaryEmailAddresses = append(
		profile.SecondaryEmailAddresses,
		uniqueEmailAddresses...)

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "id",
		Value:          profile.ID,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to parse user profile as firebase snapshot: %v", err),
		)
	}
	if len(docs) == 0 {
		return exceptions.InternalServerError(fmt.Errorf("user profile not found"))
	}
	updateCommand := &UpdateCommand{
		CollectionName: fr.GetUserProfileCollectionName(),
		ID:             docs[0].Ref.ID,
		Data:           profile,
	}
	err = fr.FirestoreClient.Update(ctx, updateCommand)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to update user profile secondary email address: %v", err),
		)
	}
	return nil
}

// UpdateSuspended updates the suspend attribute of the profile that matches the id
func (fr *Repository) UpdateSuspended(ctx context.Context, id string, status bool) error {
	ctx, span := tracer.Start(ctx, "UpdateSuspended")
	defer span.End()

	profile, err := fr.GetUserProfileByID(ctx, id, true)
	if err != nil {
		utils.RecordSpanError(span, err)
		// this is a wrapped error. No need to wrap it again
		return err
	}
	profile.Suspended = status

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "id",
		Value:          profile.ID,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(err)
	}
	if len(docs) == 0 {
		return exceptions.InternalServerError(fmt.Errorf("user profile not found"))
	}
	updateCommand := &UpdateCommand{
		CollectionName: fr.GetUserProfileCollectionName(),
		ID:             docs[0].Ref.ID,
		Data:           profile,
	}
	err = fr.FirestoreClient.Update(ctx, updateCommand)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(err)
	}

	return nil

}

// UpdatePhotoUploadID updates the photoUploadID attribute of the profile that matches the id
func (fr *Repository) UpdatePhotoUploadID(ctx context.Context, id string, uploadID string) error {
	ctx, span := tracer.Start(ctx, "UpdatePhotoUploadID")
	defer span.End()

	profile, err := fr.GetUserProfileByID(ctx, id, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		// this is a wrapped error. No need to wrap it again
		return err
	}
	profile.PhotoUploadID = uploadID

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "id",
		Value:          profile.ID,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to parse user profile as firebase snapshot: %v", err),
		)
	}
	if len(docs) == 0 {
		return exceptions.InternalServerError(fmt.Errorf("user profile not found"))
	}
	updateCommand := &UpdateCommand{
		CollectionName: fr.GetUserProfileCollectionName(),
		ID:             docs[0].Ref.ID,
		Data:           profile,
	}
	err = fr.FirestoreClient.Update(ctx, updateCommand)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to update user profile photo upload id: %v", err),
		)
	}

	return nil
}

// UpdatePushTokens updates the pushTokens attribute of the profile that matches the id. This function does a hard reset instead of prior
// matching
func (fr *Repository) UpdatePushTokens(ctx context.Context, id string, pushTokens []string) error {
	ctx, span := tracer.Start(ctx, "UpdatePushTokens")
	defer span.End()

	profile, err := fr.GetUserProfileByID(ctx, id, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		// this is a wrapped error. No need to wrap it again
		return err
	}

	profile.PushTokens = pushTokens

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "id",
		Value:          profile.ID,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to parse user profile as firebase snapshot: %v", err),
		)
	}
	if len(docs) == 0 {
		return exceptions.InternalServerError(fmt.Errorf("user profile not found"))
	}
	updateCommand := &UpdateCommand{
		CollectionName: fr.GetUserProfileCollectionName(),
		ID:             docs[0].Ref.ID,
		Data:           profile,
	}
	err = fr.FirestoreClient.Update(ctx, updateCommand)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to update user profile push tokens: %v", err),
		)
	}
	return nil
}

// UpdatePermissions update the permissions of the user profile
func (fr *Repository) UpdatePermissions(
	ctx context.Context,
	id string,
	perms []profileutils.PermissionType,
) error {
	ctx, span := tracer.Start(ctx, "UpdatePermissions")
	defer span.End()

	profile, err := fr.GetUserProfileByID(ctx, id, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		// this is a wrapped error. No need to wrap it again
		return err
	}

	// Removes duplicate permissions from array
	// Used for cleaning existing records
	profile.Permissions = utils.UniquePermissionsArray(profile.Permissions)

	newPerms := []profileutils.PermissionType{}
	// Check if has perms
	if len(profile.Permissions) >= 1 {
		// copy the existing perms
		newPerms = append(newPerms, profile.Permissions...)

		for _, perm := range perms {
			// add permission if it doesn't exist
			if !profile.HasPermission(perm) {
				newPerms = append(newPerms, perm)
			}
		}

	} else {
		newPerms = append(newPerms, perms...)
	}

	profile.Permissions = newPerms

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "id",
		Value:          profile.ID,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to parse user profile as firebase snapshot: %v", err),
		)
	}
	if len(docs) == 0 {
		return exceptions.InternalServerError(fmt.Errorf("user profile not found"))
	}
	updateCommand := &UpdateCommand{
		CollectionName: fr.GetUserProfileCollectionName(),
		ID:             docs[0].Ref.ID,
		Data:           profile,
	}
	err = fr.FirestoreClient.Update(ctx, updateCommand)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to update user profile permissions: %v", err),
		)
	}
	return nil

}

// UpdateRole update the permissions of the user profile
func (fr *Repository) UpdateRole(ctx context.Context, id string, role profileutils.RoleType) error {
	ctx, span := tracer.Start(ctx, "UpdateRole")
	defer span.End()

	profile, err := fr.GetUserProfileByID(ctx, id, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		// this is a wrapped error. No need to wrap it again
		return err
	}

	profile.Role = role
	profile.Permissions = role.Permissions()

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "id",
		Value:          profile.ID,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to parse user profile as firebase snapshot: %v", err),
		)
	}
	if len(docs) == 0 {
		return exceptions.InternalServerError(fmt.Errorf("user profile not found"))
	}
	updateCommand := &UpdateCommand{
		CollectionName: fr.GetUserProfileCollectionName(),
		ID:             docs[0].Ref.ID,
		Data:           profile,
	}
	err = fr.FirestoreClient.Update(ctx, updateCommand)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to update user role and permissions: %v", err),
		)
	}
	return nil

}

// UpdateFavNavActions update the permissions of the user profile
func (fr *Repository) UpdateFavNavActions(
	ctx context.Context,
	id string,
	favActions []string,
) error {
	profile, err := fr.GetUserProfileByID(ctx, id, false)
	if err != nil {
		// this is a wrapped error. No need to wrap it again
		return err
	}

	profile.FavNavActions = favActions

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "id",
		Value:          profile.ID,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		return exceptions.InternalServerError(
			fmt.Errorf("unable to parse user profile as firebase snapshot: %v", err),
		)
	}
	if len(docs) == 0 {
		return exceptions.InternalServerError(fmt.Errorf("user profile not found"))
	}
	updateCommand := &UpdateCommand{
		CollectionName: fr.GetUserProfileCollectionName(),
		ID:             docs[0].Ref.ID,
		Data:           profile,
	}
	err = fr.FirestoreClient.Update(ctx, updateCommand)
	if err != nil {
		return exceptions.InternalServerError(
			fmt.Errorf("unable to update user favorite actions: %v", err),
		)
	}
	return nil
}

// UpdateBioData updates the biodate of the profile that matches the id
func (fr *Repository) UpdateBioData(
	ctx context.Context,
	id string,
	data profileutils.BioData,
) error {
	ctx, span := tracer.Start(ctx, "UpdateBioData")
	defer span.End()

	profile, err := fr.GetUserProfileByID(ctx, id, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		// this is a wrapped error. No need to wrap it again
		return err
	}

	profile.UserBioData.FirstName = func(pr *profileutils.UserProfile, dt profileutils.BioData) *string {
		if dt.FirstName != nil {
			return dt.FirstName
		}
		return pr.UserBioData.FirstName
	}(
		profile,
		data,
	)
	profile.UserBioData.LastName = func(pr *profileutils.UserProfile, dt profileutils.BioData) *string {
		if dt.LastName != nil {
			return dt.LastName
		}
		return pr.UserBioData.LastName
	}(
		profile,
		data,
	)
	profile.UserBioData.Gender = func(pr *profileutils.UserProfile, dt profileutils.BioData) enumutils.Gender {
		if dt.Gender.String() != "" {
			return dt.Gender
		}
		return pr.UserBioData.Gender
	}(
		profile,
		data,
	)
	profile.UserBioData.DateOfBirth = func(pr *profileutils.UserProfile, dt profileutils.BioData) *scalarutils.Date {
		if dt.DateOfBirth != nil {
			return dt.DateOfBirth
		}

		return pr.UserBioData.DateOfBirth
	}(
		profile,
		data,
	)
	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "id",
		Value:          profile.ID,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to parse user profile as firebase snapshot: %v", err),
		)
	}
	if len(docs) == 0 {
		return exceptions.InternalServerError(fmt.Errorf("user profile not found"))
	}
	updateCommand := &UpdateCommand{
		CollectionName: fr.GetUserProfileCollectionName(),
		ID:             docs[0].Ref.ID,
		Data:           profile,
	}
	err = fr.FirestoreClient.Update(ctx, updateCommand)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to update user profile bio data: %v", err),
		)
	}
	return nil
}

// UpdateVerifiedIdentifiers adds a UID to a user profile during login if it does not exist
func (fr *Repository) UpdateVerifiedIdentifiers(
	ctx context.Context,
	id string,
	identifiers []profileutils.VerifiedIdentifier,
) error {
	ctx, span := tracer.Start(ctx, "UpdateVerifiedIdentifiers")
	defer span.End()

	for _, identifier := range identifiers {
		// for each run, get the user profile. this will ensure the fetch profile always has the latest data
		profile, err := fr.GetUserProfileByID(ctx, id, false)
		if err != nil {
			utils.RecordSpanError(span, err)
			// this is a wrapped error. No need to wrap it again
			return err
		}

		if !utils.CheckIdentifierExists(profile, identifier.UID) {
			uids := profile.VerifiedIdentifiers

			uids = append(uids, identifier)

			profile.VerifiedIdentifiers = append(profile.VerifiedIdentifiers, uids...)

			query := &GetAllQuery{
				CollectionName: fr.GetUserProfileCollectionName(),
				FieldName:      "id",
				Value:          profile.ID,
				Operator:       "==",
			}
			docs, err := fr.FirestoreClient.GetAll(ctx, query)
			if err != nil {
				utils.RecordSpanError(span, err)
				return exceptions.InternalServerError(
					fmt.Errorf("unable to parse user profile as firebase snapshot: %v", err),
				)
			}
			if len(docs) == 0 {
				return exceptions.InternalServerError(fmt.Errorf("user profile not found"))
			}
			updateCommand := &UpdateCommand{
				CollectionName: fr.GetUserProfileCollectionName(),
				ID:             docs[0].Ref.ID,
				Data:           profile,
			}
			err = fr.FirestoreClient.Update(ctx, updateCommand)
			if err != nil {
				utils.RecordSpanError(span, err)
				return exceptions.InternalServerError(
					fmt.Errorf("unable to update user profile verified identifiers: %v", err),
				)
			}
			return nil

		}
	}

	return nil
}

// UpdateVerifiedUIDS adds a UID to a user profile during login if it does not exist
func (fr *Repository) UpdateVerifiedUIDS(ctx context.Context, id string, uids []string) error {
	ctx, span := tracer.Start(ctx, "UpdateVerifiedUIDS")
	defer span.End()

	for _, uid := range uids {
		// for each run, get the user profile. this will ensure the fetch profile always has the latest data
		profile, err := fr.GetUserProfileByID(ctx, id, false)
		if err != nil {
			utils.RecordSpanError(span, err)
			// this is a wrapped error. No need to wrap it again
			return err
		}

		if !converterandformatter.StringSliceContains(profile.VerifiedUIDS, uid) {
			uids := []string{}

			uids = append(uids, uid)

			profile.VerifiedUIDS = append(profile.VerifiedUIDS, uids...)

			query := &GetAllQuery{
				CollectionName: fr.GetUserProfileCollectionName(),
				FieldName:      "id",
				Value:          profile.ID,
				Operator:       "==",
			}
			docs, err := fr.FirestoreClient.GetAll(ctx, query)
			if err != nil {
				utils.RecordSpanError(span, err)
				return exceptions.InternalServerError(
					fmt.Errorf("unable to parse user profile as firebase snapshot: %v", err),
				)
			}
			if len(docs) == 0 {
				return exceptions.InternalServerError(fmt.Errorf("user profile not found"))
			}
			updateCommand := &UpdateCommand{
				CollectionName: fr.GetUserProfileCollectionName(),
				ID:             docs[0].Ref.ID,
				Data:           profile,
			}
			err = fr.FirestoreClient.Update(ctx, updateCommand)
			if err != nil {
				utils.RecordSpanError(span, err)
				return exceptions.InternalServerError(
					fmt.Errorf("unable to update user profile verified UIDS: %v", err),
				)
			}
			return nil

		}
	}

	return nil
}

// RecordPostVisitSurvey records an end of visit survey
func (fr *Repository) RecordPostVisitSurvey(
	ctx context.Context,
	input dto.PostVisitSurveyInput,
	UID string,
) error {
	ctx, span := tracer.Start(ctx, "RecordPostVisitSurvey")
	defer span.End()

	if input.LikelyToRecommend < 0 || input.LikelyToRecommend > 10 {
		return exceptions.LikelyToRecommendError(
			fmt.Errorf("the likelihood of recommending should be an int between 0 and 10"),
		)

	}
	feedback := domain.PostVisitSurvey{
		LikelyToRecommend: input.LikelyToRecommend,
		Criticism:         input.Criticism,
		Suggestions:       input.Suggestions,
		UID:               UID,
		Timestamp:         time.Now(),
	}
	command := &CreateCommand{
		CollectionName: fr.GetSurveyCollectionName(),
		Data:           feedback,
	}
	_, err := fr.FirestoreClient.Create(ctx, command)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.AddRecordError(err)

	}
	return nil
}

// SavePIN  persist the data of the newly created PIN to a datastore
func (fr *Repository) SavePIN(ctx context.Context, pin *domain.PIN) (bool, error) {
	ctx, span := tracer.Start(ctx, "SavePin")
	defer span.End()

	// persist the data to a datastore
	command := &CreateCommand{
		CollectionName: fr.GetPINsCollectionName(),
		Data:           pin,
	}
	_, err := fr.FirestoreClient.Create(ctx, command)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.AddRecordError(err)
	}
	return true, nil

}

// UpdatePIN  persist the data of the updated PIN to a datastore
func (fr *Repository) UpdatePIN(ctx context.Context, id string, pin *domain.PIN) (bool, error) {
	ctx, span := tracer.Start(ctx, "UpdatePIN")
	defer span.End()

	pinData, err := fr.GetPINByProfileID(ctx, id)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.PinNotFoundError(err)
	}
	query := &GetAllQuery{
		CollectionName: fr.GetPINsCollectionName(),
		FieldName:      "id",
		Value:          pinData.ID,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.InternalServerError(
			fmt.Errorf("unable to parse user pin as firebase snapshot: %v", err),
		)
	}
	if len(docs) == 0 {
		return false, exceptions.InternalServerError(fmt.Errorf("user pin not found"))
	}

	// Check if PIN being updated is a Temporary PIN
	if pinData.IsOTP {
		// Set New PIN flag as false
		pin.IsOTP = false
	}

	updateCommand := &UpdateCommand{
		CollectionName: fr.GetPINsCollectionName(),
		ID:             docs[0].Ref.ID,
		Data:           pin,
	}
	err = fr.FirestoreClient.Update(ctx, updateCommand)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.UpdateProfileError(err)
	}

	return true, nil

}

// ExchangeRefreshTokenForIDToken takes a custom Firebase refresh token and tries to fetch
// an ID token and returns auth credentials if successful
// Otherwise, an error is returned
func (fr Repository) ExchangeRefreshTokenForIDToken(
	ctx context.Context,
	refreshToken string,
) (*profileutils.AuthCredentialResponse, error) {
	_, span := tracer.Start(ctx, "ExchangeRefreshTokenForIDToken")
	defer span.End()

	apiKey, err := serverutils.GetEnvVar(firebasetools.FirebaseWebAPIKeyEnvVarName)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(err)
	}

	payload := dto.RefreshTokenExchangePayload{
		GrantType:    "refresh_token",
		RefreshToken: refreshToken,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(err)
	}

	url := firebaseExchangeRefreshTokenURL + apiKey
	httpClient := http.DefaultClient
	httpClient.Timeout = time.Second * firebasetools.HTTPClientTimeoutSecs
	resp, err := httpClient.Post(
		url,
		"application/json",
		bytes.NewReader(payloadBytes),
	)

	defer firebasetools.CloseRespBody(resp)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(err)
	}

	if resp.StatusCode != http.StatusOK {
		bs, err := ioutil.ReadAll(resp.Body)
		return nil,
			exceptions.InternalServerError(fmt.Errorf(
				"firebase HTTP error, status code %d\nBody: %s\nBody read error: %s",
				resp.StatusCode,
				string(bs),
				err,
			))
	}

	type refreshTokenResponse struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    string `json:"expires_in"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
		UserID       string `json:"user_id"`
		ProjectID    string `json:"project_id"`
	}

	var tokenResponse refreshTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(fmt.Errorf(
			"failed to decode refresh token response: %s", err,
		))
	}

	profile, err := fr.GetUserProfileByUID(ctx, tokenResponse.UserID, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(fmt.Errorf(
			"failed to retrieve user profile: %s", err,
		))
	}

	canExperiment, err := fr.CheckIfExperimentParticipant(ctx, profile.ID)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(fmt.Errorf(
			"failed to check if the logged in user is an experimental participant: %s", err,
		))
	}

	return &profileutils.AuthCredentialResponse{
		IDToken:       &tokenResponse.IDToken,
		ExpiresIn:     tokenResponse.ExpiresIn,
		RefreshToken:  tokenResponse.RefreshToken,
		UID:           tokenResponse.UserID,
		IsAdmin:       fr.CheckIfAdmin(profile),
		CanExperiment: canExperiment,
	}, nil
}

// StageProfileNudge stages nudges published from this service.
func (fr *Repository) StageProfileNudge(
	ctx context.Context,
	nudge *feedlib.Nudge,
) error {
	ctx, span := tracer.Start(ctx, "StageProfileNudge")
	defer span.End()

	command := &CreateCommand{
		CollectionName: fr.GetProfileNudgesCollectionName(),
		Data:           nudge,
	}
	_, err := fr.FirestoreClient.Create(ctx, command)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(err)
	}
	return nil
}

// FetchAdminUsers fetches all admins
func (fr *Repository) FetchAdminUsers(ctx context.Context) ([]*profileutils.UserProfile, error) {
	ctx, span := tracer.Start(ctx, "FetchAdminUsers")
	defer span.End()

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "permissions",
		Value:          profileutils.DefaultAdminPermissions,
		Operator:       "array-contains-any",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, fmt.Errorf("unable to read user profile: %w", err)
	}
	var admins []*profileutils.UserProfile
	for _, doc := range docs {
		u := &profileutils.UserProfile{}
		err = doc.DataTo(u)
		if err != nil {
			utils.RecordSpanError(span, err)
			return nil, exceptions.InternalServerError(
				fmt.Errorf("unable to read user profile: %w", err),
			)
		}
		admins = append(admins, u)
	}
	return admins, nil
}

// FetchAllUsers fetches all registered users
func (fr *Repository) FetchAllUsers(ctx context.Context, callbackURL string) {
	ctx, span := tracer.Start(ctx, "FetchAllUsers")
	defer span.End()

	rl := rate.NewLimiter(rate.Every(1*time.Second), 200) // 200 request every 1 seconds
	client := utils.NewClient(rl)

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
	}

	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		logrus.Error(err)
		return
	}

	var records []*profileutils.UserProfile

	for _, doc := range docs {
		if err != nil {
			utils.RecordSpanError(span, err)
			continue
		}

		u := &profileutils.UserProfile{}
		err = doc.DataTo(u)
		if err != nil {
			utils.RecordSpanError(span, err)
			continue
		}

		if len(u.Covers) > 0 {
			records = append(records, u)
		}
	}

	for _, record := range records {
		record := record

		go func(rcd *profileutils.UserProfile) {
			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(rcd); err != nil {
				utils.RecordSpanError(span, err)
				logrus.Error(err)
				return
			}

			req, err := http.NewRequest("POST", callbackURL, &buf)
			if err != nil {
				utils.RecordSpanError(span, err)
				logrus.Error(err)
				return
			}

			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)

			if err != nil {
				utils.RecordSpanError(span, err)
				time.Sleep(2 * time.Second)
				return
			}

			// this will never happen. But because we are defensive engineers, in the event it happens,
			// we handle it appropriately
			if resp.StatusCode == 429 {
				err = errors.New("rate limit exceeded")
				utils.RecordSpanError(span, err)
				logrus.Error(err)

				// place a timeout. This is intentional because we don't want encounter another 429 again
				time.Sleep(2 * time.Second)
				return
			}

			if resp.StatusCode != http.StatusOK {
				utils.RecordSpanError(span, err)
				logrus.Error(err)
			}

		}(record)

	}
}

// PurgeUserByPhoneNumber removes the record of a user given a phone number.
func (fr *Repository) PurgeUserByPhoneNumber(ctx context.Context, phone string) error {
	ctx, span := tracer.Start(ctx, "PurgeUserByPhoneNumber")
	defer span.End()

	profile, err := fr.GetUserProfileByPhoneNumber(ctx, phone, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(err)
	}

	// delete pin of the user
	pin, err := fr.GetPINByProfileID(ctx, profile.ID)
	if err != nil {
		utils.RecordSpanError(span, err)
		// Should not panic but allow for deletion of the profile
		log.Printf("failed to get a user pin %v", err)
	}
	// Remove user profile with or without PIN
	if pin != nil {
		query := &GetAllQuery{
			CollectionName: fr.GetPINsCollectionName(),
			FieldName:      "id",
			Value:          pin.ID,
			Operator:       "==",
		}
		if docs, err := fr.FirestoreClient.GetAll(ctx, query); err == nil {
			command := &DeleteCommand{
				CollectionName: fr.GetPINsCollectionName(),
				ID:             docs[0].Ref.ID,
			}
			if err = fr.FirestoreClient.Delete(ctx, command); err != nil {
				return exceptions.InternalServerError(err)
			}
		}
	}

	// delete the user profile
	query1 := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "id",
		Value:          profile.ID,
		Operator:       "==",
	}
	if docs, err := fr.FirestoreClient.GetAll(ctx, query1); err == nil {
		command := &DeleteCommand{
			CollectionName: fr.GetUserProfileCollectionName(),
			ID:             docs[0].Ref.ID,
		}
		if err = fr.FirestoreClient.Delete(ctx, command); err != nil {
			return exceptions.InternalServerError(err)
		}
	}

	// delete the user from firebase
	u, err := fr.FirebaseClient.GetUserByPhoneNumber(ctx, phone)
	if err == nil {
		// only run firebase delete if firebase manages to find the user. It's not fatal if firebase fails to find the user
		if err := fr.FirebaseClient.DeleteUser(ctx, u.UID); err != nil {
			return exceptions.InternalServerError(err)
		}
	}

	return nil
}

// GetOrCreatePhoneNumberUser retrieves or creates an phone number user
// account in Firebase Authentication
func (fr *Repository) GetOrCreatePhoneNumberUser(
	ctx context.Context,
	phone string,
) (*dto.CreatedUserResponse, error) {
	ctx, span := tracer.Start(ctx, "GetOrCreatePhoneNumberUser")
	defer span.End()

	user, err := fr.FirebaseClient.GetUserByPhoneNumber(
		ctx,
		phone,
	)
	if err == nil {
		return &dto.CreatedUserResponse{
			UID:         user.UID,
			DisplayName: user.DisplayName,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
			PhotoURL:    user.PhotoURL,
			ProviderID:  user.ProviderID,
		}, nil
	}

	params := (&auth.UserToCreate{}).
		PhoneNumber(phone)
	newUser, err := fr.FirebaseClient.CreateUser(
		ctx,
		params,
	)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(err)
	}
	return &dto.CreatedUserResponse{
		UID:         newUser.UID,
		DisplayName: newUser.DisplayName,
		Email:       newUser.Email,
		PhoneNumber: newUser.PhoneNumber,
		PhotoURL:    newUser.PhotoURL,
		ProviderID:  newUser.ProviderID,
	}, nil
}

// HardResetSecondaryPhoneNumbers does a hard reset of user secondary phone numbers.
// This should be called when retiring specific secondary phone number and passing in
// the new secondary phone numbers as an argument.
func (fr *Repository) HardResetSecondaryPhoneNumbers(
	ctx context.Context,
	profile *profileutils.UserProfile,
	newSecondaryPhoneNumbers []string,
) error {
	ctx, span := tracer.Start(ctx, "HardResetSecondaryPhoneNumbers")
	defer span.End()

	profile.SecondaryPhoneNumbers = newSecondaryPhoneNumbers

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "id",
		Value:          profile.ID,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to parse user profile as firebase snapshot: %v", err),
		)
	}
	if len(docs) == 0 {
		return exceptions.InternalServerError(fmt.Errorf("user profile not found"))
	}
	updateCommand := &UpdateCommand{
		CollectionName: fr.GetUserProfileCollectionName(),
		ID:             docs[0].Ref.ID,
		Data:           profile,
	}
	err = fr.FirestoreClient.Update(ctx, updateCommand)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to update user profile secondary phone numbers: %v", err),
		)
	}

	return nil
}

// HardResetSecondaryEmailAddress does a hard reset of user secondary email addresses. This should be called when retiring specific
// secondary email addresses and passing in the new secondary email address as an argument.
func (fr *Repository) HardResetSecondaryEmailAddress(
	ctx context.Context,
	profile *profileutils.UserProfile,
	newSecondaryEmails []string,
) error {
	ctx, span := tracer.Start(ctx, "HardResetSecondaryEmailAddress")
	defer span.End()

	profile.SecondaryEmailAddresses = newSecondaryEmails

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "id",
		Value:          profile.ID,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to parse user profile as firebase snapshot: %v", err),
		)
	}
	if len(docs) == 0 {
		return exceptions.InternalServerError(fmt.Errorf("user profile not found"))
	}
	updateCommand := &UpdateCommand{
		CollectionName: fr.GetUserProfileCollectionName(),
		ID:             docs[0].Ref.ID,
		Data:           profile,
	}
	err = fr.FirestoreClient.Update(ctx, updateCommand)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(
			fmt.Errorf("unable to update user profile secondary phone numbers: %v", err),
		)
	}

	return nil
}

// CheckIfExperimentParticipant check if a user has subscribed to be an experiment participant
func (fr *Repository) CheckIfExperimentParticipant(
	ctx context.Context,
	profileID string,
) (bool, error) {
	ctx, span := tracer.Start(ctx, "CheckIfExperimentParticipant")
	defer span.End()

	query := &GetAllQuery{
		CollectionName: fr.GetExperimentParticipantCollectionName(),
		FieldName:      "id",
		Value:          profileID,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.InternalServerError(
			fmt.Errorf("unable to parse user profile as firebase snapshot: %v", err),
		)
	}

	if len(docs) == 0 {
		return false, nil
	}
	return true, nil
}

// AddUserAsExperimentParticipant adds the provided user profile as an experiment participant if does not already exist.
// this method is idempotent.
func (fr *Repository) AddUserAsExperimentParticipant(
	ctx context.Context,
	profile *profileutils.UserProfile,
) (bool, error) {
	ctx, span := tracer.Start(ctx, "AddUserAsExperimentParticipant")
	defer span.End()

	exists, err := fr.CheckIfExperimentParticipant(ctx, profile.ID)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, err
	}

	if !exists {
		createCommand := &CreateCommand{
			CollectionName: fr.GetExperimentParticipantCollectionName(),
			Data:           profile,
		}
		_, err = fr.FirestoreClient.Create(ctx, createCommand)
		if err != nil {
			utils.RecordSpanError(span, err)
			return false, exceptions.InternalServerError(
				fmt.Errorf(
					"unable to add user profile of ID %v in experiment_participant: %v",
					profile.ID,
					err,
				),
			)
		}
		return true, nil
	}
	// the user already exists as an experiment participant
	return true, nil

}

// RemoveUserAsExperimentParticipant removes the provide user profile as an experiment participant. This methold does not check
// for existence before deletion since non-existence is relatively equivalent to a removal
func (fr *Repository) RemoveUserAsExperimentParticipant(
	ctx context.Context,
	profile *profileutils.UserProfile,
) (bool, error) {
	ctx, span := tracer.Start(ctx, "RemoveUserAsExperimentParticipant")
	defer span.End()

	// fetch the document References
	query := &GetAllQuery{
		CollectionName: fr.GetExperimentParticipantCollectionName(),
		FieldName:      "id",
		Value:          profile.ID,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.InternalServerError(
			fmt.Errorf("unable to parse user profile as firebase snapshot: %v", err),
		)
	}
	// means the document was removed or does not exist
	if len(docs) == 0 {
		return true, nil
	}
	deleteCommand := &DeleteCommand{
		CollectionName: fr.GetExperimentParticipantCollectionName(),
		ID:             docs[0].Ref.ID,
	}
	err = fr.FirestoreClient.Delete(ctx, deleteCommand)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.InternalServerError(
			fmt.Errorf(
				"unable to remove user profile of ID %v from experiment_participant: %v",
				profile.ID,
				err,
			),
		)
	}

	return true, nil
}

// UpdateAddresses persists a user's home or work address information to the database
func (fr *Repository) UpdateAddresses(
	ctx context.Context,
	id string,
	address profileutils.Address,
	addressType enumutils.AddressType,
) error {
	ctx, span := tracer.Start(ctx, "UpdateAddresses")
	defer span.End()

	profile, err := fr.GetUserProfileByID(ctx, id, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		return err
	}

	switch addressType {
	case enumutils.AddressTypeHome:
		{
			profile.HomeAddress = &address
		}
	case enumutils.AddressTypeWork:
		{
			profile.WorkAddress = &address
		}
	default:
		return exceptions.WrongEnumTypeError(addressType.String())
	}

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "id",
		Value:          profile.ID,
		Operator:       "==",
	}
	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(err)
	}
	updateCommand := &UpdateCommand{
		CollectionName: fr.GetUserProfileCollectionName(),
		ID:             docs[0].Ref.ID,
		Data:           profile,
	}
	err = fr.FirestoreClient.Update(ctx, updateCommand)
	if err != nil {
		utils.RecordSpanError(span, err)
		return exceptions.InternalServerError(err)
	}
	return nil
}

// GetUserCommunicationsSettings fetches the communication settings of a specific user.
func (fr *Repository) GetUserCommunicationsSettings(
	ctx context.Context,
	profileID string,
) (*profileutils.UserCommunicationsSetting, error) {
	ctx, span := tracer.Start(ctx, "GetUserCommunicationsSettings")
	defer span.End()

	query := &GetAllQuery{
		CollectionName: fr.GetCommunicationsSettingsCollectionName(),
		FieldName:      "profileID",
		Value:          profileID,
		Operator:       "==",
	}

	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(err)
	}

	if len(docs) > 1 && serverutils.IsDebug() {
		log.Printf("> 1 communications settings with profile ID %s (count: %d)",
			profileID,
			len(docs),
		)
	}

	if len(docs) == 0 {
		return &profileutils.UserCommunicationsSetting{ProfileID: profileID}, nil
	}

	comms := &profileutils.UserCommunicationsSetting{}
	err = docs[0].DataTo(comms)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}
	return comms, nil
}

// SetUserCommunicationsSettings sets communication settings for a specific user
func (fr *Repository) SetUserCommunicationsSettings(
	ctx context.Context,
	profileID string,
	allowWhatsApp *bool,
	allowTextSms *bool,
	allowPush *bool,
	allowEmail *bool,
) (*profileutils.UserCommunicationsSetting, error) {

	ctx, span := tracer.Start(ctx, "SetUserCommunicationsSettings")
	defer span.End()

	// get the previous communications_settings
	comms, err := fr.GetUserCommunicationsSettings(ctx, profileID)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}

	setCommsSettings := profileutils.UserCommunicationsSetting{
		ID:            uuid.New().String(),
		ProfileID:     profileID,
		AllowWhatsApp: utils.MatchAndReturn(comms.AllowWhatsApp, *allowWhatsApp),
		AllowTextSMS:  utils.MatchAndReturn(comms.AllowWhatsApp, *allowTextSms),
		AllowPush:     utils.MatchAndReturn(comms.AllowWhatsApp, *allowPush),
		AllowEmail:    utils.MatchAndReturn(comms.AllowWhatsApp, *allowEmail),
	}

	createCommand := &CreateCommand{
		CollectionName: fr.GetCommunicationsSettingsCollectionName(),
		Data:           setCommsSettings,
	}
	_, err = fr.FirestoreClient.Create(ctx, createCommand)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(err)
	}

	// fetch the now set communications_settings and return it
	return fr.GetUserCommunicationsSettings(ctx, profileID)
}

// ListUserProfiles fetches all users with the specified role from the database
func (fr *Repository) ListUserProfiles(
	ctx context.Context,
	role profileutils.RoleType,
) ([]*profileutils.UserProfile, error) {
	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "role",
		Value:          role,
		Operator:       "==",
	}

	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		return nil, exceptions.InternalServerError(err)
	}

	profiles := []*profileutils.UserProfile{}

	for _, doc := range docs {
		profile := &profileutils.UserProfile{}
		err = doc.DataTo(profile)
		if err != nil {
			return nil, exceptions.InternalServerError(
				fmt.Errorf("unable to read user profile: %w", err),
			)
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// CreateRole creates a new role and persists it to the database
func (fr *Repository) CreateRole(
	ctx context.Context,
	profileID string,
	input dto.RoleInput,
) (*profileutils.Role, error) {
	ctx, span := tracer.Start(ctx, "CreateRole")
	defer span.End()

	exists, err := fr.CheckIfRoleNameExists(ctx, input.Name)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}

	if exists {
		err := fmt.Errorf("role with similar name exists:%v", input.Name)
		utils.RecordSpanError(span, err)
		return nil, err
	}

	timestamp := time.Now().In(pubsubtools.TimeLocation)

	role := profileutils.Role{
		ID:          uuid.New().String(),
		Name:        input.Name,
		Description: input.Description,
		CreatedBy:   profileID,
		Created:     timestamp,
		Active:      true,
		Scopes:      input.Scopes,
	}

	createCommad := &CreateCommand{
		CollectionName: fr.GetRolesCollectionName(),
		Data:           role,
	}

	_, err = fr.FirestoreClient.Create(ctx, createCommad)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(err)
	}

	return &role, nil
}

// GetAllRoles returns a list of all created roles
func (fr *Repository) GetAllRoles(ctx context.Context) (*[]profileutils.Role, error) {
	ctx, span := tracer.Start(ctx, "GetAllRoles")
	defer span.End()

	query := &GetAllQuery{
		CollectionName: fr.GetRolesCollectionName(),
	}

	docs, err := fr.FirestoreClient.GetAll(ctx, query)

	if err != nil {
		utils.RecordSpanError(span, err)
		err = fmt.Errorf("unable to read role")
		return nil, exceptions.InternalServerError(err)
	}

	roles := []profileutils.Role{}
	for _, doc := range docs {
		role := &profileutils.Role{}

		err := doc.DataTo(role)
		if err != nil {
			utils.RecordSpanError(span, err)
			err = fmt.Errorf("unable to read role")
			return nil, exceptions.InternalServerError(err)
		}
		roles = append(roles, *role)
	}

	return &roles, nil
}

// UpdateRoleDetails  updates the details of a role
func (fr *Repository) UpdateRoleDetails(
	ctx context.Context,
	profileID string,
	role profileutils.Role,
) (*profileutils.Role, error) {
	ctx, span := tracer.Start(ctx, "UpdateRoleDetails")
	defer span.End()

	query := &GetAllQuery{
		CollectionName: fr.GetRolesCollectionName(),
		Value:          role.ID,
		FieldName:      "id",
		Operator:       "==",
	}

	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, err
	}

	timestamp := time.Now().In(pubsubtools.TimeLocation)

	updatedRole := profileutils.Role{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Active:      role.Active,
		Scopes:      role.Scopes,
		CreatedBy:   role.CreatedBy,
		Created:     role.Created,
		UpdatedBy:   profileID,
		Updated:     timestamp,
	}

	updateCommand := &UpdateCommand{
		CollectionName: fr.GetRolesCollectionName(),
		ID:             docs[0].Ref.ID,
		Data:           updatedRole,
	}
	err = fr.FirestoreClient.Update(ctx, updateCommand)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(err)
	}

	return &updatedRole, nil
}

// GetRoleByID gets role with matching id
func (fr *Repository) GetRoleByID(ctx context.Context, roleID string) (*profileutils.Role, error) {
	ctx, span := tracer.Start(ctx, "GetRoleByID")
	defer span.End()

	query := &GetAllQuery{
		CollectionName: fr.GetRolesCollectionName(),
		FieldName:      "id",
		Value:          roleID,
		Operator:       "==",
	}

	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(err)
	}

	if len(docs) != 1 {
		err = fmt.Errorf("role not found: %v", roleID)
		utils.RecordSpanError(span, err)
		return nil, err
	}

	role := &profileutils.Role{}

	err = docs[0].DataTo(role)
	if err != nil {
		utils.RecordSpanError(span, err)
		err = fmt.Errorf("unable to read role")
		return nil, exceptions.InternalServerError(err)
	}

	return role, nil
}

// GetRolesByIDs gets all roles matching provided roleIDs if specified otherwise all roles
func (fr *Repository) GetRolesByIDs(
	ctx context.Context,
	roleIDs []string,
) (*[]profileutils.Role, error) {
	ctx, span := tracer.Start(ctx, "GetRoleByID")
	defer span.End()
	roles := []profileutils.Role{}
	// role ids provided
	for _, id := range roleIDs {
		role, err := fr.GetRoleByID(ctx, id)
		if err != nil {
			return nil, err
		}
		roles = append(roles, *role)
	}

	return &roles, nil
}

// DeleteRole removes a role permanently from the database
func (fr *Repository) DeleteRole(
	ctx context.Context,
	roleID string,
) (bool, error) {
	ctx, span := tracer.Start(ctx, "DeleteRole")
	defer span.End()

	// remove this role for all users who has it assigned
	query1 := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "roles",
		Operator:       "array-contains",
		Value:          roleID,
	}

	docs1, err := fr.FirestoreClient.GetAll(ctx, query1)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.InternalServerError(err)
	}

	for _, doc := range docs1 {
		user := &profileutils.UserProfile{}
		err = doc.DataTo(user)
		if err != nil {
			return false, fmt.Errorf("unable to parse userprofile")
		}
		newRoles := []string{}
		for _, userRole := range user.Roles {
			if userRole != roleID {
				newRoles = append(newRoles, userRole)
			}
		}
		err = fr.UpdateUserRoleIDs(ctx, user.ID, newRoles)
		if err != nil {
			utils.RecordSpanError(span, err)
			return false, exceptions.InternalServerError(err)
		}
	}

	// delete the role
	query := &GetAllQuery{
		CollectionName: fr.GetRolesCollectionName(),
		FieldName:      "id",
		Value:          roleID,
		Operator:       "==",
	}

	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.InternalServerError(err)
	}

	// means the document was removed or does not exist
	if len(docs) == 0 {
		return false, fmt.Errorf("error role does not exist")
	}
	deleteCommand := &DeleteCommand{
		CollectionName: fr.GetRolesCollectionName(),
		ID:             docs[0].Ref.ID,
	}
	err = fr.FirestoreClient.Delete(ctx, deleteCommand)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, fmt.Errorf(
			"unable to remove role of ID %v, error: %v",
			roleID,
			err,
		)
	}
	return true, nil
}

// CheckIfRoleNameExists checks if a role with a similar name exists
// Ensures unique name for each role during creation
func (fr *Repository) CheckIfRoleNameExists(ctx context.Context, name string) (bool, error) {
	ctx, span := tracer.Start(ctx, "CheckIfRoleNameExists")
	defer span.End()

	query := &GetAllQuery{
		CollectionName: fr.GetRolesCollectionName(),
		FieldName:      "name",
		Operator:       "==",
		Value:          name,
	}

	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, exceptions.InternalServerError(err)
	}

	if len(docs) == 1 {
		return true, nil
	}

	return false, nil
}

// GetUserProfilesByRoleID returns a list of user profiles with the role ID
// i.e users assigned a particular role
func (fr *Repository) GetUserProfilesByRoleID(ctx context.Context, roleID string) ([]*profileutils.UserProfile, error) {
	ctx, span := tracer.Start(ctx, "GetUserProfilesByRoleID")
	defer span.End()

	query := &GetAllQuery{
		CollectionName: fr.GetUserProfileCollectionName(),
		FieldName:      "roles",
		Operator:       "array-contains",
		Value:          roleID,
	}

	users := []*profileutils.UserProfile{}

	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(err)
	}

	for _, doc := range docs {
		user := &profileutils.UserProfile{}

		err = doc.DataTo(user)
		if err != nil {
			return nil, fmt.Errorf("unable to parse userprofile")
		}

		users = append(users, user)
	}

	return users, nil
}

// GetRoleByName retrieves a role using it's name
func (fr *Repository) GetRoleByName(ctx context.Context, roleName string) (*profileutils.Role, error) {
	ctx, span := tracer.Start(ctx, "GetRoleByName")
	defer span.End()

	query := &GetAllQuery{
		CollectionName: fr.GetRolesCollectionName(),
		FieldName:      "name",
		Operator:       "==",
		Value:          roleName,
	}

	docs, err := fr.FirestoreClient.GetAll(ctx, query)
	if err != nil {
		utils.RecordSpanError(span, err)
		return nil, exceptions.InternalServerError(err)
	}

	if len(docs) != 1 {
		err = fmt.Errorf("role with name %v not found", roleName)
		utils.RecordSpanError(span, err)
		return nil, err
	}

	role := &profileutils.Role{}

	err = docs[0].DataTo(role)
	if err != nil {
		utils.RecordSpanError(span, err)
		err = fmt.Errorf("unable to read role")
		return nil, exceptions.InternalServerError(err)
	}

	return role, nil
}

// SaveRoleRevocation records a log for a role revocation
//
// userId is the ID of the user removing a role from a user
func (fr *Repository) SaveRoleRevocation(ctx context.Context, userID string, revocation dto.RoleRevocationInput) error {
	ctx, span := tracer.Start(ctx, "SaveRoleRevocation")
	defer span.End()

	timestamp := time.Now().In(pubsubtools.TimeLocation)

	role := domain.RoleRevocationLog{
		ID:        uuid.New().String(),
		ProfileID: revocation.ProfileID,
		RoleID:    revocation.RoleID,
		Reason:    revocation.Reason,
		CreatedBy: userID,
		Created:   timestamp,
	}

	createCommad := &CreateCommand{
		CollectionName: fr.GetRolesRevocationCollectionName(),
		Data:           role,
	}

	_, err := fr.FirestoreClient.Create(ctx, createCommad)
	if err != nil {
		utils.RecordSpanError(span, err)
		return err
	}

	return nil
}

//CheckIfUserHasPermission checks if a user has the required permission
func (fr *Repository) CheckIfUserHasPermission(
	ctx context.Context,
	UID string,
	requiredPermission profileutils.Permission,
) (bool, error) {
	ctx, span := tracer.Start(ctx, "CheckIfUserHasPermission")
	defer span.End()

	userprofile, err := fr.GetUserProfileByUID(ctx, UID, false)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, err
	}

	roles, err := fr.GetRolesByIDs(ctx, userprofile.Roles)
	if err != nil {
		utils.RecordSpanError(span, err)
		return false, err
	}

	for _, role := range *roles {
		if role.Active && role.HasPermission(ctx, requiredPermission.Scope) {
			return true, nil
		}
	}

	return false, nil
}
