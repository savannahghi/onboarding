package presentation

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/mux"
	"github.com/savannahghi/firebasetools"
	"github.com/savannahghi/interserviceclient"
	"github.com/savannahghi/onboarding/pkg/onboarding/presentation/rest"
)

// SharedRoutes return REST routes shared by open/closed onboarding services
func SharedRoutes(handlers rest.HandlersInterfaces, r *mux.Router) {
	SharedUnauthenticatedRoutes(handlers, r)
	SharedAuthenticatedISCRoutes(handlers, r)
	SharedAuthenticatedRoutes(handlers, r)
}

// SharedUnauthenticatedRoutes return REST routes shared by open/closed onboarding services
func SharedUnauthenticatedRoutes(handlers rest.HandlersInterfaces, r *mux.Router) {
	// Unauthenticated routes
	r.Path("/switch_flagged_features").Methods(
		http.MethodPost,
		http.MethodOptions,
	).HandlerFunc(
		handlers.SwitchFlaggedFeaturesHandler(),
	)

	// misc routes
	r.Path("/ide").HandlerFunc(playground.Handler("GraphQL IDE", "/graphql"))
	r.Path("/health").HandlerFunc(HealthStatusCheck)

	// Admin service polling
	r.Path("/poll_services").Methods(http.MethodGet).HandlerFunc(handlers.PollServices())

	// signup routes
	r.Path("/verify_phone").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.VerifySignUpPhoneNumber())
	r.Path("/create_user_by_phone").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.CreateUserWithPhoneNumber())
	r.Path("/user_recovery_phonenumbers").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.UserRecoveryPhoneNumbers())
	r.Path("/set_primary_phonenumber").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.SetPrimaryPhoneNumber())

	// LoginByPhone routes
	r.Path("/login_by_phone").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.LoginByPhone())
	r.Path("/login_anonymous").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.LoginAnonymous())
	r.Path("/refresh_token").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.RefreshToken())

	// PIN Routes
	r.Path("/reset_pin").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.ResetPin())

	r.Path("/request_pin_reset").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.RequestPINReset())

	//OTP routes
	r.Path("/send_otp").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.SendOTP())

	r.Path("/send_retry_otp").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.SendRetryOTP())

	r.Path("/remove_user").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.RemoveUserByPhoneNumber())

	r.Path("/add_admin_permissions").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.AddAdminPermsToUser())

	r.Path("/remove_admin_permissions").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.RemoveAdminPermsToUser())
}

// SharedAuthenticatedRoutes return REST routes shared by open/closed onboarding services
func SharedAuthenticatedRoutes(handlers rest.HandlersInterfaces, r *mux.Router) {
	fc := &firebasetools.FirebaseClient{}
	firebaseApp, _ := fc.InitFirebase()

	// Authenticated routes
	rs := r.PathPrefix("/roles").Subrouter()
	rs.Use(firebasetools.AuthenticationMiddleware(firebaseApp))
	rs.Path("/create_role").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.CreateRole())
	rs.Path("/assign_role").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.AssignRole())
	rs.Path("/remove_role").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.RemoveRoleByName())

	rs.Path("/add_user_role").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.AddRoleToUser())

	rs.Path("/remove_user_role").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.RemoveRoleToUser())

}

// SharedAuthenticatedISCRoutes return ISC REST routes shared by open/closed onboarding services
func SharedAuthenticatedISCRoutes(handlers rest.HandlersInterfaces, r *mux.Router) {
	// Interservice Authenticated routes
	isc := r.PathPrefix("/internal").Subrouter()
	isc.Use(interserviceclient.InterServiceAuthenticationMiddleware())
	isc.Path("/user_profile").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.GetUserProfileByUID())
	isc.Path("/retrieve_user_profile").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.GetUserProfileByPhoneOrEmail())
	isc.Path("/contactdetails/{attribute}/").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.ProfileAttributes())
	isc.Path("/check_permission").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.CheckHasPermission())

	// Interservice Authenticated routes
	// The reason for the below endpoints to be used for interservice communication
	// is to allow for the creation and deletion of internal `test` users that can be used
	// to run tests in other services that require an authenticated user.
	// These endpoint have been used in the `Base` lib to create and delete the test users
	iscTesting := r.PathPrefix("/testing").Subrouter()
	iscTesting.Use(interserviceclient.InterServiceAuthenticationMiddleware())
	iscTesting.Path("/verify_phone").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.VerifySignUpPhoneNumber())
	iscTesting.Path("/create_user_by_phone").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.CreateUserWithPhoneNumber())
	iscTesting.Path("/login_by_phone").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.LoginByPhone())
	iscTesting.Path("/remove_user").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.RemoveUserByPhoneNumber())
	iscTesting.Path("/register_push_token").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.RegisterPushToken())
	iscTesting.Path("/add_admin_permissions").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.AddAdminPermsToUser())
	iscTesting.Path("/add_user_role").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.AddRoleToUser())
	iscTesting.Path("/remove_user_role").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.RemoveRoleToUser())
	iscTesting.Path("/update_user_profile").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.UpdateUserProfile())
	iscTesting.Path("/create_role").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.CreateRole())
	iscTesting.Path("/assign_role").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.AssignRole())
	iscTesting.Path("/remove_role").Methods(
		http.MethodPost,
		http.MethodOptions).
		HandlerFunc(handlers.RemoveRoleByName())
}
