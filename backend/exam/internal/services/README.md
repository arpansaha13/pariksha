## GetAnswerForExam, GetAnswerForEvaluation and GetAnswerEvaluationData

- `GetAnswerForExam` is used **by a participant** to fetch the answer corresponding to an `examId`, `userId` (of participant) and `questionId`.
- `GetAnswerForEvaluation` is used **by an evaluator** to fetch the answer corresponding to a `participantId` and `questionId`.
- `GetAnswerEvaluationData` is used **by an evaluator** to fetch the score_awarded of an answer corresponding to a `participantId` and `questionId`.

The separation between `GetAnswerForEvaluation` and `GetAnswerEvaluationData` is designed for efficiency during the evaluation process:

1. `GetAnswerForEvaluation` loads the answer content once:
   - The answer content is static after exam completion
   - Evaluators cannot modify the answer content
   - Answer content may be large, so fetching it once reduces network load

2. `GetAnswerEvaluationData` handles the dynamic evaluation data:
   - Focuses on `score_awarded` only
   - These fields change frequently during evaluation
   - Smaller payload size enables efficient repeated requests