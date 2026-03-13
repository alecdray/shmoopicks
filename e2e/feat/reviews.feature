Feature: Rankings and Reviews

  Users can rate albums on a 0–10 scale and write free-text review notes.
  Ratings can be entered directly or computed via a guided questionnaire.
  Each score maps to an opinionated label.

  Scenario: Rating modal opens to the confirmation form
    Given a logged-in user with an album in their library
    When they open the rating modal for that album
    Then the rating confirmation form is shown

  Scenario: Navigating to the questionnaire from the confirmation form
    Given a logged-in user with the rating confirmation form open
    When they click the questionnaire button
    Then the questionnaire form is shown

  Scenario: Completing the questionnaire produces a score
    Given a logged-in user with the rating questionnaire open
    When they answer all questions and click Calculate Rating
    Then the rating confirmation form is shown with a computed score

  Scenario: Saving a rating
    Given a logged-in user on the rating confirmation form
    When they enter a score between 0 and 10 and click Lock in
    Then the modal closes

  Scenario: Deleting a rating
    Given a logged-in user on the rating confirmation form for a rated album
    When they click the delete rating button
    Then the modal closes

  Scenario: Saving review notes
    Given a logged-in user with the notes modal open
    When they type some text and click Save Notes
    Then the modal closes
