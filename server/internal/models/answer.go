package models

type Answer struct {
	ID                int `gorm:"primaryKey"`
	ExamParticipantID int
	QuestionID        int
	Answer            string `gorm:"type:text"`
	ScoreAwarded      int
	Comments          string          `gorm:"type:text"`
	ExamParticipant   ExamParticipant `gorm:"foreignKey:ExamParticipantID"`
	Question          Question        `gorm:"foreignKey:QuestionID"`
}

func (Answer) TableName() string {
	return "answers"
}
