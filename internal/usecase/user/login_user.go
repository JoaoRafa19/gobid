package user

import (
	"context"
	"github.com/JoaoRafa19/gobid/internal/validator"
)

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r LoginUserRequest) Valid(ctx context.Context) validator.Evaluator {
	var eval validator.Evaluator = make(validator.Evaluator)

	eval.CheckField(validator.Matches(r.Email, validator.EmailRX), "email", "must be a valid email address")
	eval.CheckField(validator.NotBlank(r.Password), "password", "password is required")

	return eval
}
