package mocks

import "pariksha/common/pkg/proto"

var questionMap = map[int64]*proto.QuestionResponse{
	1: {
		Id:          1,
		Hash:        "q_hash_1",
		RawQuestion: []byte("What is the capital of France?"),
		Type:        proto.QuestionType_MCQ,
		TestCases:   []*proto.CodingQuestionTestCase{},
	},
	2: {
		Id:          2,
		Hash:        "q_hash_2",
		RawQuestion: []byte("Solve x in the equation 2x + 3 = 7."),
		Type:        proto.QuestionType_SUBJECTIVE,
		TestCases:   []*proto.CodingQuestionTestCase{},
	},
	3: {
		Id:          3,
		Hash:        "q_hash_3",
		RawQuestion: []byte("Write a function to reverse a string."),
		Type:        proto.QuestionType_CODING,
		TestCases: []*proto.CodingQuestionTestCase{
			{Inputs: []string{"hello"}, Output: "olleh"},
			{Inputs: []string{"world"}, Output: "dlrow"},
		},
	},
	4: {
		Id:          4,
		Hash:        "q_hash_4",
		RawQuestion: []byte("Who discovered gravity?"),
		Type:        proto.QuestionType_MCQ,
		TestCases:   []*proto.CodingQuestionTestCase{},
	},
	5: {
		Id:          5,
		Hash:        "q_hash_5",
		RawQuestion: []byte("Explain the significance of photosynthesis."),
		Type:        proto.QuestionType_SUBJECTIVE,
		TestCases:   []*proto.CodingQuestionTestCase{},
	},
	6: {
		Id:          6,
		Hash:        "q_hash_6",
		RawQuestion: []byte("Write a program to calculate factorial."),
		Type:        proto.QuestionType_CODING,
		TestCases: []*proto.CodingQuestionTestCase{
			{Inputs: []string{"5"}, Output: "120"},
			{Inputs: []string{"0"}, Output: "1"},
		},
	},
	7: {
		Id:          7,
		Hash:        "q_hash_7",
		RawQuestion: []byte("What is the boiling point of water in Celsius?"),
		Type:        proto.QuestionType_MCQ,
		TestCases:   []*proto.CodingQuestionTestCase{},
	},
	8: {
		Id:          8,
		Hash:        "q_hash_8",
		RawQuestion: []byte("Describe the process of mitosis."),
		Type:        proto.QuestionType_SUBJECTIVE,
		TestCases:   []*proto.CodingQuestionTestCase{},
	},
	9: {
		Id:          9,
		Hash:        "q_hash_9",
		RawQuestion: []byte("Write code to check if a number is prime."),
		Type:        proto.QuestionType_CODING,
		TestCases: []*proto.CodingQuestionTestCase{
			{Inputs: []string{"7"}, Output: "true"},
			{Inputs: []string{"8"}, Output: "false"},
		},
	},
	10: {
		Id:          10,
		Hash:        "q_hash_10",
		RawQuestion: []byte("Name the three states of matter."),
		Type:        proto.QuestionType_MCQ,
		TestCases:   []*proto.CodingQuestionTestCase{},
	},
}

var categoryMap = map[int64]*proto.CategoryResponse{
	1:  {Id: 1, Name: "Mathematics"},
	2:  {Id: 2, Name: "Physics"},
	3:  {Id: 3, Name: "Chemistry"},
	4:  {Id: 4, Name: "Biology"},
	5:  {Id: 5, Name: "Computer Science"},
	6:  {Id: 6, Name: "English Language"},
	7:  {Id: 7, Name: "Logical Reasoning"},
	8:  {Id: 8, Name: "General Knowledge"},
	9:  {Id: 9, Name: "Economics"},
	10: {Id: 10, Name: "Environmental Science"},
}
