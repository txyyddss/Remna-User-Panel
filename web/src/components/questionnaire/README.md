# Questionnaire components

- `QuestionnairePage.vue` loads the active member questionnaire, distinguishes a failed initial request from an empty catalog, and offers a localized retry. Existing questionnaire data stays visible if a later request fails.
- `QuestionnaireAccessPanel.vue` displays the reward, validation code, and external form action.

Questionnaire controls use copy, open, and refresh feedback according to their action meaning.

`QuestionnaireAccessPanel.vue` owns access-code presence so grant state can change without replaying the route.
