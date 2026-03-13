Feature: Tagging

  Users can apply custom tags to albums to organise their library. Tags can
  belong to optional tag groups or stand alone. Tags are managed via a modal
  with an autocomplete input and chip display.

  Scenario: Tags modal opens from the album detail page
    Given a logged-in user on an album detail page
    When they click the tags button
    Then the tags modal opens

  Scenario: Typing a tag name adds a chip
    Given a logged-in user with the tags modal open
    When they type a tag name and press Enter
    Then the tag appears as a chip in the modal

  Scenario: Saving tags closes the modal
    Given a logged-in user with the tags modal open
    When they click Save Tags
    Then the modal closes
