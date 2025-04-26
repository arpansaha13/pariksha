## `locked` column in questions and question_categories

When an exam is created using a paper, we should use the same questions that were in the paper at the time of exam creation. Even if the paper is updated later, the questions in the exam should not change.

We could ensure this by making a copy of the paper at the time of exam creation. But this will lead to a lot of duplicated data every time an exam is created. Also, if the paper has many questions, then making a copy will become an expensive task.

So instead we follow the approach of locking the questions/categories when a exam is created. When this question/category is later updated, it will not update in the same row, but create a new entry for the updated question. Hence the exam will still be referencing to the old question.

The questions and categories used in a exam are stored in a separate `exam_questions` and `exam_categories` table.

## `exam_questions` table

The `exam_questions` table stores the `category_id` and `order` only. Remaining question data will be fetched from `questions` table.

This is because, if a **locked** category is updated, a new category will be created and all questions belonging to that category needs to point to that new category in the paper. But in the exam those questions still need to be under the old category. So in the context of exams, the `category_id` from the `exam_questions` table will be used, and in the context of paper the `category_id` from the `questions` table will be used.

Similarly, in the context of exams, the `order` from the `exam_questions` table will be used, and in the context of paper the `order` from the `questions` table will be used.
