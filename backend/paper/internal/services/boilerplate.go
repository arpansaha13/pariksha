package services

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pariksha/common/pkg/proto"
	"pariksha/paper/internal/repositories"
)

type Boilerplate struct {
	boilerplateRepo *repositories.Boilerplate
}

func NewBoilerplate(repo *repositories.Boilerplate) *Boilerplate {
	return &Boilerplate{boilerplateRepo: repo}
}

// GetBoilerplate gets boilerplate code for a question and language
func (s *Boilerplate) GetBoilerplate(questionHash string, languageID int32) (*proto.GetBoilerplateResponse, error) {
	boilerplate, err := s.boilerplateRepo.GetByQuestionAndLanguage(questionHash, languageID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "boilerplate not found")
	}

	return &proto.GetBoilerplateResponse{
		Code: boilerplate.Code,
	}, nil
}
