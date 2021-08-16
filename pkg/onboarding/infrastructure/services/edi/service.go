package edi

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/extension"
	"github.com/savannahghi/onboarding/pkg/onboarding/domain"
	"github.com/savannahghi/onboarding/pkg/onboarding/repository"
	"github.com/segmentio/ksuid"
	"gitlab.slade360emr.com/go/apiclient"
)

// internal apis definitions
const (
	// LinkCover ISC endpoint to link user cover
	LinkCover = "internal/link_cover"

	// CoverLinkingStatusCompleted ...
	CoverLinkingStatusCompleted = "coverlinking completed"

	getSladerDataEndpoint = "internal/slader_data?%s"
)

// ServiceEdi defines the business logic required to interact with EDI
type ServiceEdi interface {
	LinkCover(
		ctx context.Context,
		phoneNumber string,
		uid string,
		pushToken []string,
	) (*http.Response, error)

	GetSladerData(ctx context.Context, phoneNumber string) (*[]apiclient.MarketingData, error)

	LinkEDIMemberCover(
		ctx context.Context,
		phoneNumber string,
		membernumber string,
		payersladecode int,
	) (*http.Response, error)

	CreateCoverLinkingRequest(
		ctx context.Context,
		phoneNumber string,
		membernumber string,
		payersladecode int,
		errorMessage string,
	) (*dto.CoverLinkingNotificationPayload, error)
}

// ServiceEDIImpl represents EDI usecases
type ServiceEDIImpl struct {
	EdiExt               extension.ISCClientExtension
	onboardingRepository repository.OnboardingRepository
}

// NewEdiService returns a new instance of edi implementations
func NewEdiService(
	edi extension.ISCClientExtension,
	r repository.OnboardingRepository,
) ServiceEdi {
	return &ServiceEDIImpl{
		EdiExt:               edi,
		onboardingRepository: r,
	}
}

// LinkCover calls the `EDI` service to link a cover to a converted, verified slade member user profile.
func (e *ServiceEDIImpl) LinkCover(
	ctx context.Context,
	phoneNumber string,
	uid string,
	pushToken []string,
) (*http.Response, error) {
	userMarketingData, err := e.GetSladerData(ctx, phoneNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to query the user's marketing details :%w", err)
	}

	for _, userData := range *userMarketingData {
		if userMarketingData != nil {
			sladeCode, err := strconv.Atoi(userData.PayerSladeCode)
			if err != nil {
				return nil, fmt.Errorf("failed to convert slade code to an int: %w", err)
			}
			payload := dto.CoverInput{
				PayerSladeCode: sladeCode,
				MemberNumber:   userData.MemberNumber,
				UID:            uid,
				PushToken:      pushToken,
			}

			resp, err := e.EdiExt.MakeRequest(
				ctx,
				http.MethodPost,
				LinkCover,
				payload,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to make an edi request for coverlinking: %w", err)
			}

			if resp.StatusCode != http.StatusOK {
				dataResponse, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read request body")
				}

				data := map[string]interface{}{}
				err = json.Unmarshal(dataResponse, &data)
				if err != nil {
					return nil, fmt.Errorf("bad data returned")
				}

				// If the response returned has an error, store the details in a collection
				// This makes it possible for the staff to review
				errorMessage, ok := data["error"]
				if ok {
					errMessage := errorMessage.(string)
					if !strings.Contains(errMessage, "cover already exists") {
						_, err := e.CreateCoverLinkingRequest(
							ctx,
							phoneNumber,
							userData.MemberNumber,
							sladeCode,
							errMessage,
						)
						if err != nil {
							return nil, err
						}
					}
				}
			}
			currentTime := time.Now()
			coverLinkingEvent := &dto.CoverLinkingEvent{
				ID:                    uuid.NewString(),
				CoverLinkingEventTime: &currentTime,
				CoverStatus:           CoverLinkingStatusCompleted,
				MemberNumber:          userData.MemberNumber,
				PhoneNumber:           userData.Phone,
			}

			if _, err := e.onboardingRepository.SaveCoverAutolinkingEvents(ctx, coverLinkingEvent); err != nil {
				log.Printf("failed to save coverlinking `completed` event: %v", err)
			}
			return resp, nil
		}
	}
	return nil, nil
}

// GetSladerData calls the `edi service` to fetch the details of a particular slader
// It searches by the phoneNumber
func (e *ServiceEDIImpl) GetSladerData(ctx context.Context, phoneNumber string) (*[]apiclient.MarketingData, error) {
	params := url.Values{}
	params.Add("phoneNumber", phoneNumber)

	resp, err := e.EdiExt.MakeRequest(
		ctx,
		http.MethodGet,
		fmt.Sprintf(
			getSladerDataEndpoint,
			params.Encode(),
		),
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get slader data:%w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to get data, with status code %v", resp.StatusCode)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to convert response to string: %v", err)
	}

	var sladerData []apiclient.MarketingData
	err = json.Unmarshal(data, &sladerData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal slader data: %v", err)
	}
	return &sladerData, nil
}

// LinkEDIMemberCover calls the `edi` service to autolink a cover for a slader who gets
// text messages from EDI and clicks on the provided link to download Be.Well.
// After the slader creates an account, a cover should be automatically appended to the profile.
func (e *ServiceEDIImpl) LinkEDIMemberCover(
	ctx context.Context,
	phoneNumber string,
	membernumber string,
	payersladecode int,
) (*http.Response, error) {
	userProfile, err := e.onboardingRepository.GetUserProfileByPhoneNumber(ctx, phoneNumber, false)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch a user profile: %w", err)
	}

	payload := dto.CoverInput{
		PayerSladeCode: payersladecode,
		MemberNumber:   membernumber,
		UID:            userProfile.VerifiedUIDS[0],
		PushToken:      userProfile.PushTokens,
	}

	resp, err := e.EdiExt.MakeRequest(
		ctx,
		http.MethodPost,
		LinkCover,
		payload,
	)
	if err != nil {
		return nil, fmt.Errorf("the error is %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		dataResponse, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body")
		}

		data := map[string]interface{}{}
		err = json.Unmarshal(dataResponse, &data)
		if err != nil {
			return nil, fmt.Errorf("bad data returned")
		}

		// If the response returned has an error, store the details in a collection
		// This makes it possible for the staff to review
		errorMessage, ok := data["error"]
		if ok {
			errMessage := errorMessage.(string)
			if !strings.Contains(errMessage, "cover already exists") {
				_, err := e.CreateCoverLinkingRequest(
					ctx,
					phoneNumber,
					membernumber,
					payersladecode,
					errMessage,
				)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return resp, nil
}

// CreateCoverLinkingRequest creates a cover linking request in the event that
// automatically linking a cover to a user's profile fails
func (e *ServiceEDIImpl) CreateCoverLinkingRequest(
	ctx context.Context,
	phoneNumber string,
	membernumber string,
	payersladecode int,
	errorMessage string,
) (*dto.CoverLinkingNotificationPayload, error) {
	userProfile, err := e.onboardingRepository.GetUserProfileByPhoneNumber(ctx, phoneNumber, false)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user profile: %w", err)
	}

	coverNotificationPayload := &dto.CoverLinkingNotificationPayload{
		ID:             ksuid.New().String(),
		TimeStamp:      time.Now(),
		Read:           false,
		PayerSladeCode: payersladecode,
		MemberNumber:   membernumber,
		State:          domain.CoverLinkingRequestPending,
		FirstName:      userProfile.UserBioData.FirstName,
		LastName:       userProfile.UserBioData.LastName,
		PhoneNumber:    phoneNumber,
		ErrorMessage:   errorMessage,
	}

	err = e.onboardingRepository.SaveCoverLinkingNotification(ctx, coverNotificationPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to save cover linking notification: %w", err)
	}

	// TODO: Send an alert to the ADMIN
	return coverNotificationPayload, nil
}
