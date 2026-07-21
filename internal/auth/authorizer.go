package auth

import (
	"fmt"

	"github.com/casbin/casbin/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func New(model, policy string) *Authorizer {
	enforcer, _ := casbin.NewEnforcer(model, policy)
	return &Authorizer{
		enforcer: enforcer,
	}
}

type Authorizer struct {
	enforcer *casbin.Enforcer
}

func (a *Authorizer) Authorize(subject, object, action string) error {
	_, err := a.enforcer.Enforce(subject, object, action)

	if err != nil {
		msg := fmt.Sprintf("%s not permitted to %s to %s", subject, action, object)

		st := status.New(codes.PermissionDenied, msg)
		return st.Err()
	}

	return nil
}
