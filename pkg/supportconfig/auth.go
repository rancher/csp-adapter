package supportconfig

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	authv1 "k8s.io/api/authentication/v1"
	v1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	authnv1 "k8s.io/client-go/kubernetes/typed/authentication/v1"
	authzv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

const (
	unauthenticatedError = "unauthenticated \n"
	authHeader           = "Authorization"
)

type AuthHandler struct {
	subjectReview authzv1.SubjectAccessReviewInterface
	tokenReview   authnv1.TokenReviewInterface
}

// NewAuthHandler is an Auth handler for specifically the supportconfig endpoint
func NewAuthHandler(sar authzv1.SubjectAccessReviewInterface, tr authnv1.TokenReviewInterface) *AuthHandler {
	auth := &AuthHandler{
		subjectReview: sar,
		tokenReview:   tr,
	}
	return auth
}

func (a *AuthHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userInfo, err := a.authenticate(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			if err.Error() != unauthenticatedError {
				logrus.Errorf("error authenticating user %s", err)
			} else {
				_, err = w.Write([]byte(err.Error()))
				if err != nil {
					logrus.Errorf("error writing response string in auth %s", err.Error())
				}
			}
			// don't go any further if we got an error during auth
			return
		}
		err = a.authorize(r, userInfo)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			_, err = w.Write([]byte(err.Error()))
			if err != nil {
				logrus.Errorf("error writing response string in auth %s", err.Error())
			}
			return
		}
		// if we didn't get any errors while authing, it's safe to call the next function
		next.ServeHTTP(w, r)
	})
}

// authenticate issues a token review for the K8s api - basically you need to have `Authorization: Bearer` set like you
// would for the K8s api. Right now we only have one endpoint so doing authn and authz at the same time makes sense.
// however, if this changes, this should be broken out into its own middleware
func (a *AuthHandler) authenticate(r *http.Request) (*authv1.UserInfo, error) {
	token, err := getTokenFromHeader(r)
	if err != nil {
		return nil, err
	}
	resp, err := a.tokenReview.Create(r.Context(), &authv1.TokenReview{
		Spec: authv1.TokenReviewSpec{
			Token: token,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	// if we aren't authenticated, don't attempt to extract user info
	if !resp.Status.Authenticated {
		return nil, fmt.Errorf(unauthenticatedError)
	}
	return &resp.Status.User, nil
}

// authorize checks if a user can get a support config by seeing if they can do "retrieve" on
// "supportconfigs.managment.cattle.io". As of now, this isn't a real CRD - more a way to do auth in a simple fashion.
// Intent is that only the admin or a user with these specific perms will be able to get the support configs
func (a *AuthHandler) authorize(r *http.Request, userInfo *authv1.UserInfo) error {
	//TODO: Real Resource?
	resp, err := a.subjectReview.Create(r.Context(), &v1.SubjectAccessReview{
		Spec: v1.SubjectAccessReviewSpec{
			ResourceAttributes: &v1.ResourceAttributes{
				Verb:     "retrieve",
				Version:  "v1",
				Resource: "supportconfigs",
				Group:    "management.cattle.io",
			},
			User:   userInfo.Username,
			Groups: userInfo.Groups,
			Extra:  toExtra(userInfo.Extra),
			UID:    userInfo.UID,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	if !resp.Status.Allowed {
		// if we aren't allowed, return an error with the reason
		return fmt.Errorf("%s", resp.Status.Reason)
	}
	return nil
}

// getTokenFromHeader finds the converts the header "Authorization: Bearer MyTokenHere" -> MyTokenHere, and returns an
// error if it fails to do so
func getTokenFromHeader(r *http.Request) (string, error) {
	headerVal := r.Header.Get(authHeader)
	if headerVal == "" {
		return "", fmt.Errorf("%s not found in headers", authHeader)
	}
	headerVal = strings.TrimSpace(headerVal)
	splitHeader := strings.Split(headerVal, " ")
	if len(splitHeader) != 2 {
		return "", fmt.Errorf("header value for %s malformed", authHeader)
	}
	// TODO: Check value[0] and ensure it is "Bearer"?
	return splitHeader[1], nil
}

// toExtra converts from one type of ExtraValue to another. I dislike this, but it does seem to be necessary
func toExtra(extra map[string]authv1.ExtraValue) map[string]v1.ExtraValue {
	result := map[string]v1.ExtraValue{}
	for k, v := range extra {
		result[k] = v1.ExtraValue(v)
	}
	return result
}
