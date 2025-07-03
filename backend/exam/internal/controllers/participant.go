package controllers

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"pariksha/common/pkg/proto"
	"pariksha/exam/internal/services"
)

type Participant struct {
	participantSvc *services.Participant
}

func NewParticipant(s *services.Participant) *Participant {
	return &Participant{
		participantSvc: s,
	}
}

// HandleGetExamParticipants handles retrieving all participants for an exam
func (c *Participant) HandleGetExamParticipants(ctx context.Context, req *proto.ExamRequest) (*proto.ParticipantList, error) {
	return c.participantSvc.GetExamParticipants(ctx, req)
}

// HandleAddExamParticipant handles adding a new participant to an exam
func (c *Participant) HandleAddExamParticipant(ctx context.Context, req *proto.AddParticipantRequest) (*proto.ParticipantResponse, error) {
	return c.participantSvc.AddExamParticipant(ctx, req)
}

// HandleRemoveExamParticipant handles removing a participant from an exam
func (c *Participant) HandleRemoveExamParticipant(ctx context.Context, req *proto.RemoveParticipantRequest) (*emptypb.Empty, error) {
	return c.participantSvc.RemoveExamParticipant(ctx, req)
}

// HandleGetExamParticipant handles retrieving participant data for the current user
func (c *Participant) HandleGetExamParticipant(ctx context.Context, req *proto.GetExamParticipantRequest) (*proto.GetExamParticipantResponse, error) {
	return c.participantSvc.GetExamParticipant(ctx, req)
}

// HandleGetParticipantById handles retrieving participant data by ID
func (c *Participant) HandleGetParticipantById(ctx context.Context, req *proto.ParticipantRequest) (*proto.ParticipantResponse, error) {
	return c.participantSvc.GetParticipantById(ctx, req)
}
