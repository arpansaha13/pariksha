## `locked` column in questions and question_categories

When an exam is created using a paper, we should use the same questions that were in the paper at the time of exam creation. Even if the paper is updated later, the questions in the exam should not change.

We could ensure this by making a copy of the paper at the time of exam creation. But this will lead to a lot of duplicated data every time an exam is created. Also, if the paper has many questions, then making a copy will become an expensive task.

So instead we follow the approach of locking the questions/categories when a exam is created. When this question/category is later updated, it will not update in the same row, but create a new entry for the updated question. Hence the exam will still be referencing to the old question.