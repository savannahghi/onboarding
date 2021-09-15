package usecases_test

import (
	"context"
	"log"
	"testing"

	"firebase.google.com/go/auth"
	"github.com/savannahghi/feedlib"
	"github.com/savannahghi/firebasetools"
	"github.com/savannahghi/interserviceclient"
	"github.com/savannahghi/onboarding/pkg/onboarding/application/dto"
	"github.com/stretchr/testify/assert"
)

func TestSwitchUserFlaggedFeature(t *testing.T) {
	primaryPhone := interserviceclient.TestUserPhoneNumber
	// clean up
	_ = testUsecase.RemoveUserByPhoneNumber(context.Background(), primaryPhone)
	otp, err := generateTestOTP(t, primaryPhone)
	log.Printf("this is the OTP %v", otp)
	assert.Nil(t, err)
	assert.NotNil(t, otp)
	pin := "4567"
	resp, err := testUsecase.CreateUserByPhone(
		context.Background(),
		&dto.SignUpInput{
			PhoneNumber: &primaryPhone,
			PIN:         &pin,
			Flavour:     feedlib.FlavourConsumer,
			OTP:         &otp.OTP,
		},
	)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Profile)
	assert.NotNil(t, resp.Profile.UserName)

	// login and assert whether the profile matches the one created earlier
	login, err := testUsecase.LoginByPhone(context.Background(), primaryPhone, pin, feedlib.FlavourConsumer)
	assert.Nil(t, err)
	assert.NotNil(t, login)
	assert.NotNil(t, login.Profile.UserName)
	assert.Equal(t, *login.Profile.UserName, *resp.Profile.UserName)
	assert.Equal(t, login.Auth.CanExperiment, false)

	res1, err := testUsecase.SwitchUserFlaggedFeatures(context.Background(), primaryPhone)
	assert.Nil(t, err)
	assert.Equal(t, res1.Status, "SUCCESS")

	// login again to verify the switch is set to true
	login1, err := testUsecase.LoginByPhone(context.Background(), primaryPhone, pin, feedlib.FlavourConsumer)
	assert.Nil(t, err)
	assert.NotNil(t, login1)
	assert.Equal(t, login1.Auth.CanExperiment, true)

	// switch again
	res2, err := testUsecase.SwitchUserFlaggedFeatures(context.Background(), primaryPhone)
	assert.Nil(t, err)
	assert.Equal(t, res2.Status, "SUCCESS")

	// login again to verify the switch is set to false
	login2, err := testUsecase.LoginByPhone(context.Background(), primaryPhone, pin, feedlib.FlavourConsumer)
	assert.Nil(t, err)
	assert.NotNil(t, login2)
	assert.Equal(t, login2.Auth.CanExperiment, false)
}

func TestUpdateUserProfileUserName(t *testing.T) {
	primaryPhone := interserviceclient.TestUserPhoneNumber
	// clean up
	_ = testUsecase.RemoveUserByPhoneNumber(context.Background(), primaryPhone)

	otp, err := generateTestOTP(t, primaryPhone)
	if err != nil {
		t.Errorf("failed to generate test OTP: %v", err)
		return
	}
	pin := "1234"
	resp, err := testUsecase.CreateUserByPhone(
		context.Background(),
		&dto.SignUpInput{
			PhoneNumber: &primaryPhone,
			PIN:         &pin,
			Flavour:     feedlib.FlavourConsumer,
			OTP:         &otp.OTP,
		},
	)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Profile)
	assert.NotNil(t, resp.Profile.UserName)

	// login and assert whether the profile matches the one created earlier
	login1, err := testUsecase.LoginByPhone(context.Background(), primaryPhone, pin, feedlib.FlavourConsumer)
	assert.Nil(t, err)
	assert.NotNil(t, login1)
	assert.NotNil(t, login1.Profile.UserName)
	assert.Equal(t, *login1.Profile.UserName, *resp.Profile.UserName)

	// create authenticated context
	ctx := context.Background()
	authCred := &auth.Token{UID: login1.Auth.UID}
	authenticatedContext := context.WithValue(
		ctx,
		firebasetools.AuthTokenContextKey,
		authCred,
	)

	err = testUsecase.UpdateUserName(authenticatedContext, "makmende1")
	assert.Nil(t, err)

	pr1, err := testUsecase.UserProfile(authenticatedContext)
	assert.Nil(t, err)
	assert.NotNil(t, pr1)
	assert.NotNil(t, pr1.UserName)
	assert.NotEqual(t, *login1.Profile.UserName, *pr1.UserName)
	assert.NotEqual(t, *resp.Profile.UserName, *pr1.UserName)

	// update the profile with the same userName. It should fail since the userName has already been taken.
	err = testUsecase.UpdateUserName(authenticatedContext, "makmende1")
	assert.NotNil(t, err)

	// update with a new unique user name
	err = testUsecase.UpdateUserName(authenticatedContext, "makmende2")
	assert.Nil(t, err)

	pr2, err := testUsecase.UserProfile(authenticatedContext)
	assert.Nil(t, err)
	assert.NotNil(t, pr2)
	assert.NotNil(t, pr2.UserName)
	assert.NotEqual(t, *login1.Profile.UserName, *pr2.UserName)
	assert.NotEqual(t, *resp.Profile.UserName, *pr2.UserName)
	assert.NotEqual(t, *pr1.UserName, *pr2.UserName)

}

// func TestSetPhoneAsPrimary(t *testing.T) {

// 	ctx := context.Background()
// 	f, _, err := InitializeTestNewOnboarding(ctx)
// 	if err != nil {
// 		t.Errorf("failed to initialize new FCM: %v", err)
// 	}

// 	primaryPhone := interserviceclient.TestUserPhoneNumber
// 	secondaryPhone := interserviceclient.TestUserPhoneNumberWithPin

// 	// clean up
// 	_ = f.RemoveUserByPhoneNumber(ctx, primaryPhone)
// 	_ = s.RemoveUserByPhoneNumber(ctx, secondaryPhone)

// 	otp, err := generateTestOTP(t, primaryPhone)
// 	if err != nil {
// 		t.Errorf("failed to generate test OTP: %v", err)
// 		return
// 	}

// 	pin := "1234"

// 	resp, err := s.CreateUserByPhone(
// 		ctx,
// 		&dto.SignUpInput{
// 			PhoneNumber: &primaryPhone,
// 			PIN:         &pin,
// 			Flavour:     feedlib.FlavourConsumer,
// 			OTP:         &otp.OTP,
// 		},
// 	)
// 	if err != nil {
// 		t.Errorf("failed to create a user by phone")
// 		return
// 	}

// 	if resp == nil {
// 		t.Error("nil user response returned")
// 		return
// 	}

// 	login1, err := s.LoginByPhone(ctx, primaryPhone, pin, feedlib.FlavourConsumer)
// 	if err != nil {
// 		t.Errorf("an error occurred while logging in by phone")
// 		return
// 	}

// 	if login1 == nil {
// 		t.Errorf("nil response returned")
// 		return
// 	}

// 	// create authenticated context
// 	authCred := &auth.Token{UID: login1.Auth.UID}
// 	authenticatedContext := context.WithValue(
// 		ctx,
// 		firebasetools.AuthTokenContextKey,
// 		authCred,
// 	)

// 	// try to login with secondaryPhone. This should fail because secondaryPhone != primaryPhone
// 	login2, err := s.LoginByPhone(ctx, secondaryPhone, pin, feedlib.FlavourConsumer)
// 	if err == nil {
// 		t.Errorf("expected an error :%v", err)
// 		return
// 	}

// 	if login2 != nil {
// 		t.Errorf("the response was not expected")
// 		return
// 	}

// 	// add a secondary phone number to the user
// 	err = s.UpdateSecondaryPhoneNumbers(authenticatedContext, []string{secondaryPhone})
// 	if err != nil {
// 		t.Errorf("failed to add a secondary number to the user")
// 		return
// 	}

// 	pr, err := s.UserProfile(authenticatedContext)
// 	if err != nil {
// 		t.Errorf("failed to retrieve the profile of the logged in user")
// 		return
// 	}

// 	if pr == nil {
// 		t.Errorf("nil response returned")
// 		return
// 	}
// 	// // check if the length of secondary number == 1
// 	if len(pr.SecondaryPhoneNumbers) != 1 {
// 		t.Errorf("expected the value to be equal to 1")
// 		return
// 	}

// 	// login to add assert the secondary phone number has been added
// 	login3, err := s.LoginByPhone(ctx, primaryPhone, pin, feedlib.FlavourConsumer)
// 	if err != nil {
// 		t.Errorf("expected an error :%v", err)
// 		return
// 	}

// 	if login3 == nil {
// 		t.Errorf("the response was not expected")
// 		return
// 	}

// 	// // check if the length of secondary number == 1
// 	if len(login3.Profile.SecondaryPhoneNumbers) != 1 {
// 		t.Errorf("expected the value to be equal to 1")
// 		return
// 	}

// 	// send otp to the secondary phone number we intend to make primary
// 	testAppID := uuid.New().String()
// 	otpResp, err := infrastructure.GenerateAndSendOTP(context.Background(), secondaryPhone, &testAppID)
// 	if err != nil {
// 		t.Errorf("unable to send generate and send otp :%v", err)
// 		return
// 	}

// 	if otpResp == nil {
// 		t.Errorf("unexpected response")
// 		return
// 	}

// 	// set the old secondary phone number as the new primary phone number
// 	setResp, err := s.SetPhoneAsPrimary(ctx, secondaryPhone, otpResp.OTP)
// 	if err != nil {
// 		t.Errorf("failed to set phone as primary: %v", err)
// 		return
// 	}

// 	if setResp == false {
// 		t.Errorf("unexpected response")
// 		return
// 	}

// 	// login with the old primary phone number. This should fail
// 	login4, err := s.LoginByPhone(ctx, primaryPhone, pin, feedlib.FlavourConsumer)
// 	if err == nil {
// 		t.Errorf("unexpected error occurred! :%v", err)
// 		return
// 	}

// 	if login4 != nil {
// 		t.Errorf("unexpected error occurred! Expected this to fail")
// 		return
// 	}

// 	// login with the new primary phone number. This should not fail. Assert that the primary phone number
// 	// is the new one and the secondary phone slice contains the old primary phone number.
// 	login5, err := s.LoginByPhone(ctx, secondaryPhone, pin, feedlib.FlavourConsumer)
// 	if err != nil {
// 		t.Errorf("failed to login by phone :%v", err)
// 		return
// 	}

// 	if login5 == nil {
// 		t.Errorf("the response was not expected")
// 		return
// 	}

// 	if secondaryPhone != *login5.Profile.PrimaryPhone {
// 		t.Errorf("expected %v and %v to be equal", secondaryPhone, *login5.Profile.PrimaryPhone)
// 		return
// 	}

// 	_, exist := utils.FindItem(login5.Profile.SecondaryPhoneNumbers, primaryPhone)
// 	if !exist {
// 		t.Errorf("the secondary phonenumber slice %v, does not contain %v",
// 			login5.Profile.SecondaryPhoneNumbers,
// 			primaryPhone,
// 		)
// 		return
// 	}

// 	// clean up
// 	_ = s.RemoveUserByPhoneNumber(ctx, secondaryPhone)
// }

// func TestAddSecondaryPhoneNumbers(t *testing.T) {
// 	s := testUsecase
// 	primaryPhone := interserviceclient.TestUserPhoneNumber
// 	secondaryPhone1 := interserviceclient.TestUserPhoneNumberWithPin
// 	secondaryPhone2 := "+25712345690"
// 	secondaryPhone3 := "+25710375600"

// 	// clean up
// 	_ = s.RemoveUserByPhoneNumber(context.Background(), primaryPhone)

// 	otp, err := generateTestOTP(t, primaryPhone)
// 	if err != nil {
// 		t.Errorf("failed to generate test OTP: %v", err)
// 		return
// 	}
// 	pin := "1234"
// 	resp, err := s.CreateUserByPhone(
// 		context.Background(),
// 		&dto.SignUpInput{
// 			PhoneNumber: &primaryPhone,
// 			PIN:         &pin,
// 			Flavour:     feedlib.FlavourConsumer,
// 			OTP:         &otp.OTP,
// 		},
// 	)
// 	if err != nil {
// 		t.Errorf("failed to create a user by phone")
// 		return
// 	}

// 	if resp == nil {
// 		t.Error("nil user response returned")
// 		return
// 	}

// 	if resp.Profile == nil {
// 		t.Error("nil profile response returned")
// 		return
// 	}

// 	login1, err := s.LoginByPhone(context.Background(), primaryPhone, pin, feedlib.FlavourConsumer)
// 	if err != nil {
// 		t.Errorf("an error occurred while logging in by phone :%v", err)
// 		return
// 	}

// 	if login1 == nil {
// 		t.Errorf("nil response returned")
// 		return
// 	}

// 	// create authenticated context
// 	ctx := context.Background()
// 	authCred := &auth.Token{UID: login1.Auth.UID}
// 	authenticatedContext := context.WithValue(
// 		ctx,
// 		firebasetools.AuthTokenContextKey,
// 		authCred,
// 	)

// 	// add the first secondary phone number
// 	err = s.UpdateSecondaryPhoneNumbers(authenticatedContext, []string{secondaryPhone1})
// 	if err != nil {
// 		t.Errorf("failed to add secondary phonenumber :%v", err)
// 		return
// 	}

// 	userProfile, err := s.UserProfile(authenticatedContext)
// 	if err != nil {
// 		t.Errorf("failed to retrieve the profile of the logged in user :%v", err)
// 		return
// 	}

// 	if userProfile == nil {
// 		t.Errorf("nil response returned")
// 		return
// 	}

// 	// check if the length of secondary number == 1
// 	if len(userProfile.SecondaryPhoneNumbers) != 1 {
// 		t.Errorf("expected the value to be equal to %v",
// 			len(userProfile.SecondaryPhoneNumbers),
// 		)
// 		return
// 	}

// 	// try adding secondaryPhone1 again. this should fail because secondaryPhone1 already exists
// 	err = s.UpdateSecondaryPhoneNumbers(authenticatedContext, []string{secondaryPhone1})
// 	if err == nil {
// 		t.Errorf("an error %v was expected", err)
// 		return
// 	}

// 	// add the second secondary phone number
// 	err = s.UpdateSecondaryPhoneNumbers(authenticatedContext, []string{secondaryPhone2})
// 	if err != nil {
// 		t.Errorf("failed to add secondary phonenumber :%v", err)
// 		return
// 	}

// 	userProfile, err = s.UserProfile(authenticatedContext)
// 	if err != nil {
// 		t.Errorf("failed to retrieve the profile of the logged in user :%v", err)
// 		return
// 	}

// 	if userProfile == nil {
// 		t.Errorf("nil response returned")
// 		return
// 	}

// 	// check if the length of secondary number == 2
// 	if len(userProfile.SecondaryPhoneNumbers) != 2 {
// 		t.Errorf("expected the value to be equal to %v",
// 			len(userProfile.SecondaryPhoneNumbers),
// 		)
// 		return
// 	}

// 	// try adding secondaryPhone2 again. this should fail because secondaryPhone2 already exists
// 	err = s.UpdateSecondaryPhoneNumbers(authenticatedContext, []string{secondaryPhone2})
// 	if err == nil {
// 		t.Errorf("an error %v was expected", err)
// 		return
// 	}

// 	// add the third secondary phone number
// 	err = s.UpdateSecondaryPhoneNumbers(authenticatedContext, []string{secondaryPhone3})
// 	if err != nil {
// 		t.Errorf("failed to add secondary phonenumber :%v", err)
// 		return
// 	}

// 	userProfile, err = s.UserProfile(authenticatedContext)
// 	if err != nil {
// 		t.Errorf("failed to retrieve the profile of the logged in user :%v", err)
// 		return
// 	}

// 	if userProfile == nil {
// 		t.Errorf("nil response returned")
// 		return
// 	}

// 	// check if the length of secondary number == 3
// 	if len(userProfile.SecondaryPhoneNumbers) != 3 {
// 		t.Errorf("expected the value to be equal to %v",
// 			len(userProfile.SecondaryPhoneNumbers),
// 		)
// 		return
// 	}

// 	// try adding secondaryPhone3 again. this should fail because secondaryPhone3 already exists
// 	err = s.UpdateSecondaryPhoneNumbers(authenticatedContext, []string{secondaryPhone3})
// 	if err == nil {
// 		t.Errorf("an error %v was expected", err)
// 		return
// 	}

// 	// try to login with each secondary phone number. This should fail
// 	login2, err := s.LoginByPhone(context.Background(), secondaryPhone1, pin, feedlib.FlavourConsumer)
// 	if err == nil {
// 		t.Errorf("an error %v was expected ", err)
// 		return
// 	}

// 	if login2 != nil {
// 		t.Errorf("an unexpected error occurred :%v", err)
// 	}

// 	login3, err := s.LoginByPhone(context.Background(), secondaryPhone2, pin, feedlib.FlavourConsumer)
// 	if err == nil {
// 		t.Errorf("an error %v was expected ", err)
// 		return
// 	}

// 	if login3 != nil {
// 		t.Errorf("an unexpected error occurred :%v", err)
// 	}

// 	login4, err := s.LoginByPhone(context.Background(), secondaryPhone3, pin, feedlib.FlavourConsumer)
// 	if err == nil {
// 		t.Errorf("an error %v was expected ", err)
// 		return
// 	}

// 	if login4 != nil {
// 		t.Errorf("an unexpected error occurred :%v", err)
// 	}
// }

// func TestAddSecondaryEmailAddress(t *testing.T) {
// 	s := testUsecase
// 	primaryPhone := interserviceclient.TestUserPhoneNumber
// 	primaryEmail := "test@bewell.co.ke"
// 	secondaryemail1 := "user1@gmail.com"
// 	secondaryemail2 := "user2@gmail.com"
// 	secondaryemail3 := "user3@gmail.com"

// 	// clean up
// 	_ = s.RemoveUserByPhoneNumber(context.Background(), primaryPhone)

// 	otp, err := generateTestOTP(t, primaryPhone)
// 	if err != nil {
// 		t.Errorf("failed to generate test OTP: %v", err)
// 		return
// 	}
// 	pin := "1234"
// 	resp, err := s.CreateUserByPhone(
// 		context.Background(),
// 		&dto.SignUpInput{
// 			PhoneNumber: &primaryPhone,
// 			PIN:         &pin,
// 			Flavour:     feedlib.FlavourConsumer,
// 			OTP:         &otp.OTP,
// 		},
// 	)
// 	if err != nil {
// 		t.Errorf("failed to create a user by phone")
// 		return
// 	}

// 	if resp == nil {
// 		t.Error("nil user response returned")
// 		return
// 	}

// 	if resp.Profile == nil {
// 		t.Error("nil profile response returned")
// 		return
// 	}

// 	login1, err := s.LoginByPhone(context.Background(), primaryPhone, pin, feedlib.FlavourConsumer)
// 	if err != nil {
// 		t.Errorf("an error occurred while logging in by phone :%v", err)
// 		return
// 	}

// 	if login1 == nil {
// 		t.Errorf("nil response returned")
// 		return
// 	}

// 	// create authenticated context
// 	ctx := context.Background()
// 	authCred := &auth.Token{UID: login1.Auth.UID}
// 	authenticatedContext := context.WithValue(
// 		ctx,
// 		firebasetools.AuthTokenContextKey,
// 		authCred,
// 	)

// 	// try adding a secondary email address. This should fail because the profile does not have a primary email first
// 	err = s.UpdateSecondaryEmailAddresses(authenticatedContext, []string{secondaryemail1})
// 	if err == nil {
// 		t.Errorf("expected an error: %v", err)
// 		return
// 	}

// 	// add the profile's primary email address. This is necessary. primary email must first exist before adding secondary emails
// 	err = s.UpdatePrimaryEmailAddress(authenticatedContext, primaryEmail)
// 	if err != nil {
// 		t.Errorf("failed to add a primary email: %v", err)
// 		return
// 	}

// 	err = s.UpdateSecondaryEmailAddresses(authenticatedContext, []string{secondaryemail1})
// 	if err != nil {
// 		t.Errorf("failed to add secondary email: %v", err)
// 		return
// 	}

// 	userProfile, err := s.UserProfile(authenticatedContext)
// 	if err != nil {
// 		t.Errorf("failed to retrieve the profile of the logged in user :%v", err)
// 		return
// 	}

// 	if userProfile == nil {
// 		t.Errorf("nil response returned")
// 		return
// 	}
// 	// check if the length of secondary email == 1
// 	if len(userProfile.SecondaryEmailAddresses) != 1 {
// 		t.Errorf("expected the value to be equal to %v",
// 			len(userProfile.SecondaryEmailAddresses),
// 		)
// 		return
// 	}

// 	// try adding secondaryemail1 again since secondaryemail1 is already in use
// 	err = s.UpdateSecondaryEmailAddresses(authenticatedContext, []string{secondaryemail1})
// 	if err == nil {
// 		t.Errorf("an error %v was expected", err)
// 		return
// 	}

// 	// now add secondaryemail2
// 	err = s.UpdateSecondaryEmailAddresses(authenticatedContext, []string{secondaryemail2})
// 	if err != nil {
// 		t.Errorf("failed to add secondary email: %v", err)
// 		return
// 	}

// 	userProfile, err = s.UserProfile(authenticatedContext)
// 	if err != nil {
// 		t.Errorf("failed to retrieve the profile of the logged in user :%v", err)
// 		return
// 	}

// 	if userProfile == nil {
// 		t.Errorf("nil response returned")
// 		return
// 	}
// 	// check if the length of secondary email == 2
// 	if len(userProfile.SecondaryEmailAddresses) != 2 {
// 		t.Errorf("expected the value to be equal to %v",
// 			len(userProfile.SecondaryEmailAddresses),
// 		)
// 		return
// 	}

// 	// try adding secondaryemail2 again since secondaryemail1 is already in use
// 	err = s.UpdateSecondaryEmailAddresses(authenticatedContext, []string{secondaryemail2})
// 	if err == nil {
// 		t.Errorf("an error %v was expected", err)
// 		return
// 	}

// 	// now add secondaryemail3
// 	err = s.UpdateSecondaryEmailAddresses(authenticatedContext, []string{secondaryemail3})
// 	if err != nil {
// 		t.Errorf("failed to add secondary email: %v", err)
// 		return
// 	}

// 	userProfile, err = s.UserProfile(authenticatedContext)
// 	if err != nil {
// 		t.Errorf("failed to retrieve the profile of the logged in user :%v", err)
// 		return
// 	}

// 	if userProfile == nil {
// 		t.Errorf("nil response returned")
// 		return
// 	}
// 	// check if the length of secondary email == 3
// 	if len(userProfile.SecondaryEmailAddresses) != 3 {
// 		t.Errorf("expected the value to be equal to %v",
// 			len(userProfile.SecondaryEmailAddresses),
// 		)
// 		return
// 	}
// 	// try adding secondaryemail3 again since secondaryemail3 is already in use
// 	err = s.UpdateSecondaryEmailAddresses(authenticatedContext, []string{secondaryemail3})
// 	if err == nil {
// 		t.Errorf("an error %v was expected", err)
// 		return
// 	}

// }

// func TestUpdateUserProfilePushTokens(t *testing.T) {
// 	s := testUsecase
// 	primaryPhone := interserviceclient.TestUserPhoneNumber
// 	// clean up
// 	_ = s.RemoveUserByPhoneNumber(context.Background(), primaryPhone)

// 	otp, err := generateTestOTP(t, primaryPhone)
// 	if err != nil {
// 		t.Errorf("failed to generate test OTP: %v", err)
// 		return
// 	}
// 	pin := "1234"
// 	resp, err := s.CreateUserByPhone(
// 		context.Background(),
// 		&dto.SignUpInput{
// 			PhoneNumber: &primaryPhone,
// 			PIN:         &pin,
// 			Flavour:     feedlib.FlavourConsumer,
// 			OTP:         &otp.OTP,
// 		},
// 	)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, resp)
// 	assert.NotNil(t, resp.Profile)

// 	login1, err := s.LoginByPhone(context.Background(), primaryPhone, pin, feedlib.FlavourConsumer)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, login1)

// 	// create authenticated context
// 	ctx := context.Background()
// 	authCred := &auth.Token{UID: login1.Auth.UID}
// 	authenticatedContext := context.WithValue(
// 		ctx,
// 		firebasetools.AuthTokenContextKey,
// 		authCred,
// 	)

// 	err = s.UpdatePushTokens(context.Background(), "token1", false)
// 	assert.NotNil(t, err)

// 	err = s.UpdatePushTokens(authenticatedContext, "token1", false)
// 	assert.Nil(t, err)

// 	pr, err := s.UserProfile(authenticatedContext)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, pr)
// 	assert.Equal(t, 1, len(pr.PushTokens))

// 	err = s.UpdatePushTokens(authenticatedContext, "token2", false)
// 	assert.Nil(t, err)

// 	pr, err = s.UserProfile(authenticatedContext)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, pr)
// 	assert.Equal(t, 1, len(pr.PushTokens))

// 	err = s.UpdatePushTokens(authenticatedContext, "token3", false)
// 	assert.Nil(t, err)

// 	pr, err = s.UserProfile(authenticatedContext)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, pr)
// 	assert.Equal(t, 1, len(pr.PushTokens))

// 	// remove the token and assert new length
// 	err = s.UpdatePushTokens(context.Background(), "token2", true)
// 	assert.NotNil(t, err)

// 	err = s.UpdatePushTokens(authenticatedContext, "token2", true)
// 	assert.Nil(t, err)

// 	pr, err = s.UserProfile(authenticatedContext)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, pr)
// 	assert.Equal(t, 1, len(pr.PushTokens))

// 	err = s.UpdatePushTokens(authenticatedContext, "token1", true)
// 	assert.Nil(t, err)

// 	pr, err = s.UserProfile(authenticatedContext)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, pr)
// 	assert.Equal(t, 1, len(pr.PushTokens))
// }

// func TestCheckPhoneExists(t *testing.T) {
// 	s := testUsecase

// 	phone := interserviceclient.TestUserPhoneNumber

// 	// remove user then signup user with the phone number then run phone number check
// 	// ignore the error since it is of no consequence to us
// 	_ = s.RemoveUserByPhoneNumber(context.Background(), phone)

// 	otp, err := generateTestOTP(t, phone)
// 	if err != nil {
// 		t.Errorf("failed to generate test OTP: %v", err)
// 		return
// 	}
// 	pin := interserviceclient.TestUserPin
// 	resp, err := s.CreateUserByPhone(
// 		context.Background(),
// 		&dto.SignUpInput{
// 			PhoneNumber: &phone,
// 			PIN:         &pin,
// 			Flavour:     feedlib.FlavourConsumer,
// 			OTP:         &otp.OTP,
// 		},
// 	)

// 	assert.Nil(t, err)
// 	assert.NotNil(t, resp)

// 	resp2, err2 := s.CheckPhoneExists(context.Background(), phone)
// 	assert.Nil(t, err2)
// 	assert.NotNil(t, resp2)
// 	assert.Equal(t, true, resp2)

// 	// clean up
// 	_ = s.RemoveUserByPhoneNumber(context.Background(), phone)
// }

// func TestGetUserProfileByUID(t *testing.T) {
// 	s := testUsecase
// 	primaryPhone := interserviceclient.TestUserPhoneNumber
// 	pin := "1234"

// 	// clean up
// 	_ = s.RemoveUserByPhoneNumber(context.Background(), primaryPhone)

// 	otp, err := generateTestOTP(t, primaryPhone)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, otp)

// 	resp, err := s.CreateUserByPhone(
// 		context.Background(),
// 		&dto.SignUpInput{
// 			PhoneNumber: &primaryPhone,
// 			PIN:         &pin,
// 			Flavour:     feedlib.FlavourConsumer,
// 			OTP:         &otp.OTP,
// 		},
// 	)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, resp)
// 	assert.NotNil(t, resp.Profile)

// 	login1, err := s.LoginByPhone(context.Background(), primaryPhone, pin, feedlib.FlavourConsumer)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, login1)

// 	// create authenticated context
// 	ctx := context.Background()
// 	authCred := &auth.Token{UID: login1.Auth.UID}
// 	authenticatedContext := context.WithValue(
// 		ctx,
// 		firebasetools.AuthTokenContextKey,
// 		authCred,
// 	)

// 	// fetch the user profile using UID
// 	pr, err := s.GetUserProfileByUID(authenticatedContext, login1.Auth.UID)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, pr)
// 	assert.Equal(t, login1.Profile.ID, pr.ID)
// 	assert.Equal(t, login1.Profile.UserName, pr.UserName)

// 	// now fetch using an authenticated context. should not fail
// 	pr2, err := s.GetUserProfileByUID(context.Background(), login1.Auth.UID)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, pr2)
// 	assert.Equal(t, login1.Profile.ID, pr2.ID)
// 	assert.Equal(t, login1.Profile.UserName, pr2.UserName)
// }

// func TestUserProfile(t *testing.T) {
// 	s := testUsecase
// 	primaryPhone := interserviceclient.TestUserPhoneNumber
// 	pin := "1234"

// 	// clean up
// 	_ = s.RemoveUserByPhoneNumber(context.Background(), primaryPhone)

// 	otp, err := generateTestOTP(t, primaryPhone)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, otp)

// 	resp, err := s.CreateUserByPhone(
// 		context.Background(),
// 		&dto.SignUpInput{
// 			PhoneNumber: &primaryPhone,
// 			PIN:         &pin,
// 			Flavour:     feedlib.FlavourConsumer,
// 			OTP:         &otp.OTP,
// 		},
// 	)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, resp)
// 	assert.NotNil(t, resp.Profile)

// 	login1, err := s.LoginByPhone(context.Background(), primaryPhone, pin, feedlib.FlavourConsumer)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, login1)

// 	// create authenticated context
// 	ctx := context.Background()
// 	authCred := &auth.Token{UID: login1.Auth.UID}
// 	authenticatedContext := context.WithValue(
// 		ctx,
// 		firebasetools.AuthTokenContextKey,
// 		authCred,
// 	)

// 	// fetch the user profile using authenticated context
// 	pr, err := s.UserProfile(authenticatedContext)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, pr)
// 	assert.Equal(t, login1.Profile.ID, pr.ID)
// 	assert.Equal(t, login1.Profile.UserName, pr.UserName)

// 	// now fetch using an unauthenticated context. should fail
// 	pr2, err := s.UserProfile(context.Background())
// 	assert.NotNil(t, err)
// 	assert.Nil(t, pr2)

// }

// func TestGetProfileByID(t *testing.T) {
// 	s := testUsecase
// 	primaryPhone := interserviceclient.TestUserPhoneNumber
// 	pin := "1234"

// 	// clean up
// 	_ = s.RemoveUserByPhoneNumber(context.Background(), primaryPhone)

// 	otp, err := generateTestOTP(t, primaryPhone)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, otp)

// 	resp, err := s.CreateUserByPhone(
// 		context.Background(),
// 		&dto.SignUpInput{
// 			PhoneNumber: &primaryPhone,
// 			PIN:         &pin,
// 			Flavour:     feedlib.FlavourConsumer,
// 			OTP:         &otp.OTP,
// 		},
// 	)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, resp)
// 	assert.NotNil(t, resp.Profile)

// 	login1, err := s.LoginByPhone(context.Background(), primaryPhone, pin, feedlib.FlavourConsumer)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, login1)

// 	// create authenticated context
// 	ctx := context.Background()
// 	authCred := &auth.Token{UID: login1.Auth.UID}
// 	authenticatedContext := context.WithValue(
// 		ctx,
// 		firebasetools.AuthTokenContextKey,
// 		authCred,
// 	)

// 	// fetch the user profile using ID
// 	pr, err := s.GetProfileByID(authenticatedContext, &login1.Profile.ID)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, pr)
// 	assert.Equal(t, login1.Profile.ID, pr.ID)
// 	assert.Equal(t, login1.Profile.UserName, pr.UserName)

// 	// now fetch using an authenticated context. should not fail
// 	pr2, err := s.GetProfileByID(context.Background(), &login1.Profile.ID)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, pr2)
// 	assert.Equal(t, login1.Profile.ID, pr2.ID)
// 	assert.Equal(t, login1.Profile.UserName, pr2.UserName)

// }

// func TestUpdateBioData(t *testing.T) {
// 	s := testUsecase

// 	validPhoneNumber := interserviceclient.TestUserPhoneNumber
// 	validPIN := "1234"

// 	validFlavourConsumer := feedlib.FlavourConsumer

// 	// clean up
// 	_ = s.RemoveUserByPhoneNumber(context.Background(), validPhoneNumber)

// 	// send otp to the phone number to initiate registration process
// 	otp, err := generateTestOTP(t, validPhoneNumber)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, otp)

// 	// this should pass
// 	resp, err := s.CreateUserByPhone(
// 		context.Background(),
// 		&dto.SignUpInput{
// 			PhoneNumber: &validPhoneNumber,
// 			PIN:         &validPIN,
// 			Flavour:     validFlavourConsumer,
// 			OTP:         &otp.OTP,
// 		},
// 	)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, resp)
// 	assert.NotNil(t, resp.Profile)
// 	assert.Equal(t, validPhoneNumber, *resp.Profile.PrimaryPhone)
// 	assert.NotNil(t, resp.Profile.UserName)

// 	// create authenticated context
// 	ctx := context.Background()
// 	authCred := &auth.Token{UID: resp.Auth.UID}
// 	authenticatedContext := context.WithValue(
// 		ctx,
// 		firebasetools.AuthTokenContextKey,
// 		authCred,
// 	)

// 	dateOfBirth1 := scalarutils.Date{
// 		Day:   12,
// 		Year:  1998,
// 		Month: 2,
// 	}
// 	dateOfBirth2 := scalarutils.Date{
// 		Day:   12,
// 		Year:  1995,
// 		Month: 10,
// 	}

// 	firstName1 := "makmende1"
// 	lastName1 := "Omera1"
// 	firstName2 := "makmende2"
// 	lastName2 := "Omera2"

// 	justDOB := profileutils.BioData{
// 		DateOfBirth: &dateOfBirth1,
// 	}

// 	justFirstName := profileutils.BioData{
// 		FirstName: &firstName1,
// 	}

// 	justLastName := profileutils.BioData{
// 		LastName: &lastName1,
// 	}

// 	completeUserDetails := profileutils.BioData{
// 		DateOfBirth: &dateOfBirth2,
// 		FirstName:   &firstName2,
// 		LastName:    &lastName2,
// 	}

// 	// update just the date of birth
// 	err = s.UpdateBioData(authenticatedContext, justDOB)
// 	assert.Nil(t, err)

// 	// fetch and assert dob has been updated
// 	pr, err := s.UserProfile(authenticatedContext)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, pr)
// 	assert.Equal(t, justDOB.DateOfBirth, pr.UserBioData.DateOfBirth)

// 	// update just the firstname
// 	err = s.UpdateBioData(authenticatedContext, justFirstName)
// 	assert.Nil(t, err)

// 	// fetch and assert firstname has been updated
// 	pr, err = s.UserProfile(authenticatedContext)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, pr)
// 	assert.Equal(t, justFirstName.FirstName, pr.UserBioData.FirstName)

// 	// update just the lastname
// 	err = s.UpdateBioData(authenticatedContext, justLastName)
// 	assert.Nil(t, err)

// 	// fetch and assert firstname has been updated
// 	pr, err = s.UserProfile(authenticatedContext)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, pr)
// 	assert.Equal(t, justLastName.LastName, pr.UserBioData.LastName)

// 	// update with the entire update input
// 	err = s.UpdateBioData(authenticatedContext, completeUserDetails)
// 	assert.Nil(t, err)

// 	// fetch and assert dob, lastname & firstname have been updated
// 	pr, err = s.UserProfile(authenticatedContext)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, pr)
// 	assert.Equal(t, completeUserDetails.DateOfBirth, pr.UserBioData.DateOfBirth)
// 	assert.Equal(t, completeUserDetails.LastName, pr.UserBioData.LastName)
// 	assert.Equal(t, completeUserDetails.FirstName, pr.UserBioData.FirstName)

// 	assert.NotEqual(t, justDOB.DateOfBirth, pr.UserBioData.DateOfBirth)
// 	assert.NotEqual(t, justFirstName.FirstName, pr.UserBioData.LastName)
// 	assert.NotEqual(t, justLastName.LastName, pr.UserBioData.FirstName)

// 	// try update with an invalid context
// 	err = s.UpdateBioData(context.Background(), completeUserDetails)
// 	assert.NotNil(t, err)

// }

// func TestUpdatePhotoUploadID(t *testing.T) {
// 	s := testUsecase

// 	validPhoneNumber := interserviceclient.TestUserPhoneNumber
// 	validPIN := "1234"

// 	validFlavourConsumer := feedlib.FlavourConsumer

// 	// clean up
// 	_ = s.RemoveUserByPhoneNumber(context.Background(), validPhoneNumber)

// 	// send otp to the phone number to initiate registration process
// 	otp, err := generateTestOTP(t, validPhoneNumber)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, otp)

// 	// this should pass
// 	resp, err := s.CreateUserByPhone(
// 		context.Background(),
// 		&dto.SignUpInput{
// 			PhoneNumber: &validPhoneNumber,
// 			PIN:         &validPIN,
// 			Flavour:     validFlavourConsumer,
// 			OTP:         &otp.OTP,
// 		},
// 	)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, resp)
// 	assert.NotNil(t, resp.Profile)
// 	assert.Equal(t, validPhoneNumber, *resp.Profile.PrimaryPhone)
// 	assert.NotNil(t, resp.Profile.UserName)

// 	// create authenticated context
// 	ctx := context.Background()
// 	authCred := &auth.Token{UID: resp.Auth.UID}
// 	authenticatedContext := context.WithValue(
// 		ctx,
// 		firebasetools.AuthTokenContextKey,
// 		authCred,
// 	)

// 	uploadID1 := "photo-url1"
// 	uploadID2 := "photo-url2"

// 	err = s.UpdatePhotoUploadID(authenticatedContext, uploadID1)
// 	assert.Nil(t, err)

// 	// fetch and assert firstname has been updated
// 	pr, err := s.UserProfile(authenticatedContext)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, pr)
// 	assert.Equal(t, uploadID1, pr.PhotoUploadID)
// 	assert.NotEqual(t, resp.Profile.PhotoUploadID, pr.PhotoUploadID)

// 	err = s.UpdatePhotoUploadID(authenticatedContext, uploadID2)
// 	assert.Nil(t, err)

// 	// fetch and assert firstname has been updated again
// 	pr, err = s.UserProfile(authenticatedContext)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, pr)
// 	assert.Equal(t, uploadID2, pr.PhotoUploadID)
// 	assert.NotEqual(t, resp.Profile.PhotoUploadID, pr.PhotoUploadID)
// 	assert.NotEqual(t, uploadID1, pr.PhotoUploadID)
// }

// func TestUpdateSuspended(t *testing.T) {
// 	s := testUsecase

// 	validPhoneNumber := interserviceclient.TestUserPhoneNumber
// 	validPIN := "1234"

// 	validFlavourConsumer := feedlib.FlavourConsumer

// 	// clean up
// 	_ = s.RemoveUserByPhoneNumber(context.Background(), validPhoneNumber)

// 	// send otp to the phone number to initiate registration process
// 	otp, err := generateTestOTP(t, validPhoneNumber)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, otp)

// 	// this should pass
// 	resp, err := s.CreateUserByPhone(
// 		context.Background(),
// 		&dto.SignUpInput{
// 			PhoneNumber: &validPhoneNumber,
// 			PIN:         &validPIN,
// 			Flavour:     validFlavourConsumer,
// 			OTP:         &otp.OTP,
// 		},
// 	)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, resp)
// 	assert.NotNil(t, resp.Profile)
// 	assert.Equal(t, validPhoneNumber, *resp.Profile.PrimaryPhone)
// 	assert.NotNil(t, resp.Profile.UserName)

// 	// create authenticated context
// 	ctx := context.Background()
// 	authCred := &auth.Token{UID: resp.Auth.UID}
// 	authenticatedContext := context.WithValue(
// 		ctx,
// 		firebasetools.AuthTokenContextKey,
// 		authCred,
// 	)

// 	// fetch the profile and assert suspended
// 	pr, err := s.UserProfile(authenticatedContext)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, pr)
// 	assert.Equal(t, false, pr.Suspended)

// 	// now suspend the profile
// 	err = s.UpdateSuspended(authenticatedContext, true, *pr.PrimaryPhone, true)
// 	assert.Nil(t, err)

// 	// fetch the profile. this should fail because the profile has been suspended
// 	pr, err = s.UserProfile(authenticatedContext)
// 	assert.NotNil(t, err)
// 	assert.Nil(t, pr)
// }

// func TestUpdatePermissions(t *testing.T) {
// 	s := testUsecase

// 	validPhoneNumber := interserviceclient.TestUserPhoneNumber
// 	validPIN := "1234"

// 	validFlavourConsumer := feedlib.FlavourConsumer

// 	// clean up
// 	_ = s.RemoveUserByPhoneNumber(context.Background(), validPhoneNumber)

// 	// send otp to the phone number to initiate registration process
// 	otp, err := generateTestOTP(t, validPhoneNumber)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, otp)

// 	// this should pass
// 	resp, err := s.CreateUserByPhone(
// 		context.Background(),
// 		&dto.SignUpInput{
// 			PhoneNumber: &validPhoneNumber,
// 			PIN:         &validPIN,
// 			Flavour:     validFlavourConsumer,
// 			OTP:         &otp.OTP,
// 		},
// 	)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, resp)
// 	assert.NotNil(t, resp.Profile)
// 	assert.Equal(t, validPhoneNumber, *resp.Profile.PrimaryPhone)
// 	assert.NotNil(t, resp.Profile.UserName)

// 	// create authenticated context
// 	ctx := context.Background()
// 	authCred := &auth.Token{UID: resp.Auth.UID}
// 	authenticatedContext := context.WithValue(
// 		ctx,
// 		firebasetools.AuthTokenContextKey,
// 		authCred,
// 	)

// 	// fetch the profile and assert  the permissions slice is empty
// 	pr, err := s.UserProfile(authenticatedContext)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, pr)
// 	assert.Equal(t, 0, len(pr.Permissions))

// 	// now update the permissions
// 	perms := []profileutils.PermissionType{profileutils.PermissionTypeAdmin}
// 	err = s.UpdatePermissions(authenticatedContext, perms)
// 	assert.Nil(t, err)

// 	// fetch the profile and assert  the permissions slice is not empty
// 	pr, err = s.UserProfile(authenticatedContext)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, pr)
// 	assert.Equal(t, 1, len(pr.Permissions))

// 	// use unauthenticated context. should fail
// 	err = s.UpdatePermissions(context.Background(), perms)
// 	assert.NotNil(t, err)

// 	pr, err = s.UserProfile(context.Background())
// 	assert.NotNil(t, err)
// 	assert.Nil(t, pr)
// }

// func TestSetupAsExperimentParticipant(t *testing.T) {
// 	s := testUsecase

// 	validPhoneNumber := interserviceclient.TestUserPhoneNumber
// 	validPIN := "1234"

// 	validFlavourConsumer := feedlib.FlavourConsumer

// 	// clean up
// 	_ = s.RemoveUserByPhoneNumber(context.Background(), validPhoneNumber)

// 	// send otp to the phone number to initiate registration process
// 	otp, err := generateTestOTP(t, validPhoneNumber)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, otp)

// 	// this should pass
// 	resp, err := s.CreateUserByPhone(
// 		context.Background(),
// 		&dto.SignUpInput{
// 			PhoneNumber: &validPhoneNumber,
// 			PIN:         &validPIN,
// 			Flavour:     validFlavourConsumer,
// 			OTP:         &otp.OTP,
// 		},
// 	)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, resp)
// 	assert.NotNil(t, resp.Profile)
// 	assert.Equal(t, validPhoneNumber, *resp.Profile.PrimaryPhone)
// 	assert.NotNil(t, resp.Profile.UserName)
// 	// check that the currently created user can not experiment on new features
// 	assert.Equal(t, false, resp.Auth.CanExperiment)

// 	// create authenticated context
// 	ctx := context.Background()
// 	authCred := &auth.Token{UID: resp.Auth.UID}
// 	authenticatedContext := context.WithValue(
// 		ctx,
// 		firebasetools.AuthTokenContextKey,
// 		authCred,
// 	)

// 	// now add the user as an experiment participant
// 	input := true
// 	status, err := s.SetupAsExperimentParticipant(authenticatedContext, &input)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, status)
// 	assert.Equal(t, true, status)

// 	// try to add the user as an experiment participant. This should return the the same response since th method internally is idempotent
// 	status, err = s.SetupAsExperimentParticipant(authenticatedContext, &input)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, status)
// 	assert.Equal(t, true, status)

// 	// login the user and assert they can experiment on new features
// 	login1, err := s.LoginByPhone(context.Background(), validPhoneNumber, validPIN, validFlavourConsumer)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, login1)
// 	assert.Equal(t, true, login1.Auth.CanExperiment)

// 	// now remove the user as an experiment participant
// 	input2 := false
// 	status, err = s.SetupAsExperimentParticipant(authenticatedContext, &input2)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, status)
// 	assert.Equal(t, true, status)

// 	// try removing the user as an experiment participant.This should return the the same response since th method internally is idempotent
// 	status, err = s.SetupAsExperimentParticipant(authenticatedContext, &input2)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, status)
// 	assert.Equal(t, true, status)

// 	// login the user and assert they can not experiment on new features
// 	login2, err := s.LoginByPhone(context.Background(), validPhoneNumber, validPIN, validFlavourConsumer)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, login1)
// 	assert.Equal(t, false, login2.Auth.CanExperiment)
// }

// func TestMaskPhoneNumbers(t *testing.T) {
// 	s := testUsecase
// 	type args struct {
// 		phones []string
// 	}

// 	tests := []struct {
// 		name string
// 		arg  args
// 		want []string
// 	}{
// 		{
// 			name: "valid case",
// 			arg: args{
// 				phones: []string{"+254789874267"},
// 			},
// 			want: []string{"+254789***267"},
// 		},
// 		{
// 			name: "valid case < 10 digits",
// 			arg: args{
// 				phones: []string{"+2547898742"},
// 			},
// 			want: []string{"+2547***742"},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			maskedPhone := s.MaskPhoneNumbers(tt.arg.phones)
// 			if len(maskedPhone) != len(tt.want) {
// 				t.Errorf("returned masked phone number not the expected one, wanted: %v got: %v", tt.want, maskedPhone)
// 				return
// 			}

// 			for i, number := range maskedPhone {
// 				if tt.want[i] != number {
// 					t.Errorf("wanted: %v, got: %v", tt.want[i], number)
// 					return
// 				}
// 			}
// 		})
// 	}
// }

// func TestAddAddress(t *testing.T) {
// 	ctx, _, err := GetTestAuthenticatedContext(t)
// 	if err != nil {
// 		t.Errorf("failed to get test authenticated context: %v", err)
// 		return
// 	}
// 	s := testUsecase
// 	if err != nil {
// 		t.Errorf("unable to initialize test service")
// 		return
// 	}

// 	addr := dto.UserAddressInput{
// 		Latitude:  1.2,
// 		Longitude: -34.001,
// 	}

// 	type args struct {
// 		ctx         context.Context
// 		input       dto.UserAddressInput
// 		addressType enumutils.AddressType
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name: "happy:) add home address",
// 			args: args{
// 				ctx:         ctx,
// 				input:       addr,
// 				addressType: enumutils.AddressTypeHome,
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "happy:) add work address",
// 			args: args{
// 				ctx:         ctx,
// 				input:       addr,
// 				addressType: enumutils.AddressTypeWork,
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "sad:( failed to get logged in user",
// 			args: args{
// 				ctx:         context.Background(),
// 				input:       addr,
// 				addressType: enumutils.AddressTypeWork,
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			_, err := s.AddAddress(tt.args.ctx, tt.args.input, tt.args.addressType)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("ProfileUseCaseImpl.AddAddress() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 		})
// 	}
// }

// func TestGetAddresses(t *testing.T) {
// 	ctx, _, err := GetTestAuthenticatedContext(t)
// 	if err != nil {
// 		t.Errorf("failed to get test authenticated context: %v", err)
// 		return
// 	}
// 	s := testUsecase
// 	if err != nil {
// 		t.Errorf("unable to initialize test service")
// 		return
// 	}

// 	addr := dto.UserAddressInput{
// 		Latitude:  1.2,
// 		Longitude: -34.001,
// 	}

// 	_, err = s.AddAddress(ctx, addr, enumutils.AddressTypeWork)
// 	if err != nil {
// 		t.Errorf("unable to add test address")
// 		return
// 	}

// 	type args struct {
// 		ctx context.Context
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name:    "happy:) get addresses",
// 			args:    args{ctx: ctx},
// 			wantErr: false,
// 		},
// 		{
// 			name:    "sad:( failed to get logged in user",
// 			args:    args{ctx: context.Background()},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			_, err := s.GetAddresses(tt.args.ctx)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("ProfileUseCaseImpl.GetAddresses() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 		})
// 	}
// }

// func TestIntegrationGetAddresses(t *testing.T) {
// 	s := testUsecase

// 	validPhoneNumber := interserviceclient.TestUserPhoneNumber
// 	validPIN := interserviceclient.TestUserPin
// 	validFlavourConsumer := feedlib.FlavourConsumer

// 	_ = s.RemoveUserByPhoneNumber(
// 		context.Background(),
// 		validPhoneNumber,
// 	)

// 	otp, err := generateTestOTP(t, validPhoneNumber)
// 	if err != nil {
// 		t.Errorf("an error occurred: %v", err)
// 		return
// 	}

// 	resp, err := s.CreateUserByPhone(
// 		context.Background(),
// 		&dto.SignUpInput{
// 			PhoneNumber: &validPhoneNumber,
// 			PIN:         &validPIN,
// 			Flavour:     validFlavourConsumer,
// 			OTP:         &otp.OTP,
// 		},
// 	)
// 	if err != nil {
// 		t.Errorf("an error occurred: %v", err)
// 		return
// 	}
// 	if resp.Profile.HomeAddress != nil {
// 		t.Errorf("did not expect an address")
// 		return
// 	}
// 	if resp.Profile.WorkAddress != nil {
// 		t.Errorf("did not expect an address")
// 		return
// 	}

// 	// create authenticated context
// 	ctx := context.Background()
// 	authCred := &auth.Token{UID: resp.Auth.UID}
// 	authenticatedContext := context.WithValue(
// 		ctx,
// 		firebasetools.AuthTokenContextKey,
// 		authCred,
// 	)

// 	lat := -1.2
// 	long := 34.56

// 	addr, err := s.AddAddress(
// 		authenticatedContext,
// 		dto.UserAddressInput{
// 			Latitude:  lat,
// 			Longitude: long,
// 		},
// 		enumutils.AddressTypeHome,
// 	)
// 	if err != nil {
// 		t.Errorf("an error occurred: %v", err)
// 		return
// 	}
// 	if addr == nil {
// 		t.Errorf("expected an address")
// 		return
// 	}

// 	addrLat := addr.Latitude
// 	addrLong := addr.Longitude

// 	if addrLat != fmt.Sprintf("%.15f", lat) {
// 		t.Errorf("got a wrong address Latitude")
// 		return
// 	}
// 	if addrLong != fmt.Sprintf("%.15f", long) {
// 		t.Errorf("got a wrong address Longitude")
// 		return
// 	}

// 	profile, err := s.UserProfile(authenticatedContext)
// 	if err != nil {
// 		t.Errorf("an error occurred: %v", err)
// 		return
// 	}
// 	if profile == nil {
// 		t.Errorf("expected a user profile")
// 		return
// 	}

// 	if profile.HomeAddress == nil {
// 		t.Errorf("we expected an address")
// 		return
// 	}

// 	err = s.RemoveUserByPhoneNumber(
// 		authenticatedContext,
// 		validPhoneNumber,
// 	)
// 	if err != nil {
// 		t.Errorf("an error occurred: %v", err)
// 		return
// 	}
// }

// func TestRetireSecondaryPhoneNumbers(t *testing.T) {
// 	ctx, _, err := GetTestAuthenticatedContext(t)
// 	if err != nil {
// 		t.Errorf("failed to get test authenticated context: %v", err)
// 		return
// 	}
// 	p := testUsecase
// 	type args struct {
// 		phoneNumbers []string
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    bool
// 		wantErr bool
// 	}{
// 		{
// 			name: "sad :( unable to get the user profile",
// 			args: args{
// 				phoneNumbers: []string{uuid.New().String()},
// 			},
// 			want:    false,
// 			wantErr: true,
// 		},
// 		{
// 			name: "sad :( profile with no secondary phonenumbers",
// 			args: args{
// 				phoneNumbers: []string{uuid.New().String()},
// 			},
// 			want:    false,
// 			wantErr: true,
// 		},
// 		{
// 			name: "sad :( adding an already existent phone number",
// 			args: args{
// 				phoneNumbers: []string{interserviceclient.TestUserPhoneNumber},
// 			},
// 			want:    false,
// 			wantErr: true,
// 		},
// 		{
// 			name: "happy :) retire secondary phone numbers",
// 			args: args{
// 				phoneNumbers: []string{"+254700000003", "+254700000001"},
// 			},
// 			want:    true,
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if tt.name == "sad :( unable to get the user profile" {
// 				got, err := p.RetireSecondaryPhoneNumbers(context.Background(), tt.args.phoneNumbers)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("ProfileUseCaseImpl.RetireSecondaryPhoneNumbers() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if got != tt.want {
// 					t.Errorf("ProfileUseCaseImpl.RetireSecondaryPhoneNumbers() = %v, want %v", got, tt.want)
// 				}
// 			}

// 			if tt.name == "sad :( profile with no secondary phonenumbers" {
// 				got, err := p.RetireSecondaryPhoneNumbers(ctx, tt.args.phoneNumbers)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("ProfileUseCaseImpl.RetireSecondaryPhoneNumbers() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if got != tt.want {
// 					t.Errorf("ProfileUseCaseImpl.RetireSecondaryPhoneNumbers() = %v, want %v", got, tt.want)
// 				}
// 			}

// 			if tt.name == "sad :( adding an already existent phone number" {
// 				err := p.UpdateSecondaryPhoneNumbers(ctx, []string{"+254700000001"})
// 				if err != nil {
// 					t.Errorf("unable to add secondary phone numbers: %v", err)
// 					return
// 				}

// 				got, err := p.RetireSecondaryPhoneNumbers(ctx, tt.args.phoneNumbers)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("ProfileUseCaseImpl.RetireSecondaryPhoneNumbers() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if got != tt.want {
// 					t.Errorf("ProfileUseCaseImpl.RetireSecondaryPhoneNumbers() = %v, want %v", got, tt.want)
// 				}
// 				profile, err := testUsecase.UserProfile(ctx)
// 				if err != nil {
// 					t.Errorf("unable to get user profile")
// 					return
// 				}
// 				if len(profile.SecondaryPhoneNumbers) > 1 {
// 					t.Errorf("expected 1 secondary phone numbers")
// 					return
// 				}
// 			}

// 			if tt.name == "happy :) retire secondary phone numbers" {
// 				err := p.UpdateSecondaryPhoneNumbers(ctx, []string{"+254700000003"})
// 				if err != nil {
// 					t.Errorf("unable to add secondary phone numbers: %v", err)
// 					return
// 				}

// 				got, err := p.RetireSecondaryPhoneNumbers(ctx, tt.args.phoneNumbers)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("ProfileUseCaseImpl.RetireSecondaryPhoneNumbers() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if got != tt.want {
// 					t.Errorf("ProfileUseCaseImpl.RetireSecondaryPhoneNumbers() = %v, want %v", got, tt.want)
// 				}
// 				profile, err := testUsecase.UserProfile(ctx)
// 				if err != nil {
// 					t.Errorf("unable to get user profile")
// 					return
// 				}

// 				if len(profile.SecondaryPhoneNumbers) > 0 {
// 					t.Errorf("expected 0 secondary phone numbers but got: %v", len(profile.SecondaryPhoneNumbers))
// 					return
// 				}
// 			}
// 		})
// 	}
// }

// func TestRetireSecondaryEmailAddress(t *testing.T) {
// 	ctx, _, err := GetTestAuthenticatedContext(t)
// 	if err != nil {
// 		t.Errorf("failed to get test authenticated context: %v", err)
// 		return
// 	}
// 	p := testUsecase
// 	testEmail := "randommail@gmail.com"
// 	type args struct {
// 		emailAddresses []string
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    bool
// 		wantErr bool
// 	}{
// 		{
// 			name: "sad :( unable to get the user profile",
// 			args: args{
// 				emailAddresses: []string{converterandformatter.GenerateRandomEmail()},
// 			},
// 			want:    false,
// 			wantErr: true,
// 		},
// 		{
// 			name: "sad :( profile with no secondary email addresses",
// 			args: args{
// 				emailAddresses: []string{converterandformatter.GenerateRandomEmail()},
// 			},
// 			want:    false,
// 			wantErr: true,
// 		},
// 		{
// 			name: "sad :( adding an already existent email addresses",
// 			args: args{
// 				emailAddresses: []string{firebasetools.TestUserEmail},
// 			},
// 			want:    false,
// 			wantErr: true,
// 		},
// 		{
// 			name: "happy :) retire secondary email addresses",
// 			args: args{
// 				emailAddresses: []string{testEmail},
// 			},
// 			want:    true,
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if tt.name == "sad :( unable to get the user profile" {
// 				got, err := p.RetireSecondaryEmailAddress(context.Background(), tt.args.emailAddresses)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("ProfileUseCaseImpl.RetireSecondaryEmailAddress() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if got != tt.want {
// 					t.Errorf("ProfileUseCaseImpl.RetireSecondaryEmailAddress() = %v, want %v", got, tt.want)
// 				}
// 			}

// 			if tt.name == "sad :( profile with no secondary email addresses" {
// 				got, err := p.RetireSecondaryEmailAddress(ctx, tt.args.emailAddresses)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("ProfileUseCaseImpl.RetireSecondaryEmailAddress() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if got != tt.want {
// 					t.Errorf("ProfileUseCaseImpl.RetireSecondaryEmailAddress() = %v, want %v", got, tt.want)
// 				}
// 			}

// 			if tt.name == "sad :( adding an already existent email addresses" {
// 				err := p.UpdatePrimaryEmailAddress(ctx, converterandformatter.GenerateRandomEmail())
// 				if err != nil {
// 					t.Errorf("unable to set primary email address: %v", err)
// 					return
// 				}

// 				got, err := p.RetireSecondaryEmailAddress(ctx, tt.args.emailAddresses)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("ProfileUseCaseImpl.RetireSecondaryEmailAddress() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if got != tt.want {
// 					t.Errorf("ProfileUseCaseImpl.RetireSecondaryEmailAddress() = %v, want %v", got, tt.want)
// 				}
// 				profile, err := testUsecase.UserProfile(ctx)
// 				if err != nil {
// 					t.Errorf("unable to get user profile")
// 					return
// 				}
// 				if len(profile.SecondaryEmailAddresses) > 1 {
// 					t.Errorf("expected 1 secondary email addresses")
// 					return
// 				}
// 			}

// 			if tt.name == "happy :) retire secondary email addresses" {
// 				profile, err := p.UserProfile(ctx)
// 				if err != nil {
// 					t.Errorf("unable to get user profile")
// 					return
// 				}

// 				err = p.UpdatePrimaryEmailAddress(ctx, firebasetools.TestUserEmail)
// 				if err != nil {
// 					t.Errorf("unable to set primary email address: %v", err)
// 					return
// 				}

// 				time.Sleep(2 * time.Second)

// 				err = p.UpdateSecondaryEmailAddresses(ctx, []string{testEmail})
// 				if err != nil {
// 					t.Errorf("unable to set secondary email address: %v", err)
// 					return
// 				}

// 				got, err := p.RetireSecondaryEmailAddress(ctx, tt.args.emailAddresses)
// 				if (err != nil) != tt.wantErr {
// 					t.Errorf("ProfileUseCaseImpl.RetireSecondaryEmailAddress() error = %v, wantErr %v", err, tt.wantErr)
// 					return
// 				}
// 				if got != tt.want {
// 					t.Errorf("ProfileUseCaseImpl.RetireSecondaryEmailAddress() = %v, want %v", got, tt.want)
// 				}
// 				if len(profile.SecondaryEmailAddresses) > 0 {
// 					t.Errorf("expected 0 secondary email addresses but got: %v", len(profile.SecondaryEmailAddresses))
// 					return
// 				}
// 			}
// 		})
// 	}
// }

// func TestProfileUseCaseImpl_RemoveAdminPermsToUser(t *testing.T) {
// 	ctx, _, err := GetTestAuthenticatedContext(t)
// 	if err != nil {
// 		t.Errorf("failed to get test authenticated context: %v", err)
// 		return
// 	}
// 	s := testUsecase

// 	phoneNumber := interserviceclient.TestUserPhoneNumber
// 	p := testUsecase

// 	_ = s.RemoveUserByPhoneNumber(
// 		context.Background(),
// 		phoneNumber,
// 	)
// 	phoneNumberWithNoUserProfile := "+2547898742"
// 	otp, err := generateTestOTP(t, phoneNumber)
// 	if err != nil {
// 		t.Errorf("failed to generate test OTP: %v", err)
// 		return
// 	}
// 	pin := "1234"
// 	_, err = p.CreateUserByPhone(
// 		context.Background(),
// 		&dto.SignUpInput{
// 			PhoneNumber: &phoneNumber,
// 			PIN:         &pin,
// 			Flavour:     feedlib.FlavourConsumer,
// 			OTP:         &otp.OTP,
// 		},
// 	)
// 	if err != nil {
// 		t.Errorf("failed to create a user by phone")
// 		return
// 	}
// 	type args struct {
// 		ctx   context.Context
// 		phone string
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name: "happy case:) remove admin permissions ",
// 			args: args{
// 				ctx:   ctx,
// 				phone: phoneNumber,
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "sade case:) remove admin permissions",
// 			args: args{
// 				ctx:   ctx,
// 				phone: phoneNumberWithNoUserProfile,
// 			},
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := p.RemoveAdminPermsToUser(tt.args.ctx, tt.args.phone); (err != nil) != tt.wantErr {
// 				t.Errorf("ProfileUseCaseImpl.RemoveAdminPermsToUser() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestAddRoleToUser(t *testing.T) {
// 	ctx, _, err := GetTestAuthenticatedContext(t)
// 	if err != nil {
// 		t.Errorf("failed to get test authenticated context: %v", err)
// 		return
// 	}

// 	p := testUsecase

// 	type args struct {
// 		ctx   context.Context
// 		phone *string
// 		role  *profileutils.RoleType
// 	}
// 	validPhone := interserviceclient.TestUserPhoneNumber
// 	invalidPhone := "+2547000"
// 	validRole := profileutils.RoleTypeEmployee
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name: "valid: add role to user",
// 			args: args{
// 				ctx:   ctx,
// 				phone: &validPhone,
// 				role:  &validRole,
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "invalid: add role to user failed",
// 			args: args{
// 				ctx:   ctx,
// 				phone: &invalidPhone,
// 				role:  &validRole,
// 			},
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := p.AddRoleToUser(tt.args.ctx, *tt.args.phone, *tt.args.role); (err != nil) != tt.wantErr {
// 				t.Errorf("ProfileUseCaseImpl.AddRoleToUser() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestRemoveRoleToUser(t *testing.T) {
// 	ctx, _, err := GetTestAuthenticatedContext(t)
// 	if err != nil {
// 		t.Errorf("failed to get test authenticated context: %v", err)
// 		return
// 	}

// 	p := testUsecase

// 	type args struct {
// 		ctx   context.Context
// 		phone *string
// 	}
// 	validPhone := interserviceclient.TestUserPhoneNumber
// 	invalidPhone := "+2547000"
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantErr bool
// 	}{
// 		{
// 			name: "valid: add role to user",
// 			args: args{
// 				ctx:   ctx,
// 				phone: &validPhone,
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "invalid: add role to user failed",
// 			args: args{
// 				ctx:   ctx,
// 				phone: &invalidPhone,
// 			},
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := p.RemoveRoleToUser(tt.args.ctx, *tt.args.phone); (err != nil) != tt.wantErr {
// 				t.Errorf("ProfileUseCaseImpl.RemoveRoleToUser() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }
