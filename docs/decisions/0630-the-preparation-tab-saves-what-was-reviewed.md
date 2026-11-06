---
category: ux
status: accepted
approval: person
affects: [command/walk, preparation/store, tui]
builds-on: ['0340', '0350', '0520', '0570']
---

# 0630. The Preparation tab saves what was reviewed

Builds on [0340](0340-guided-preparation-save.md),
[0350](0350-guided-support-arrangements.md),
[0520](0520-preparation-screen.md), and
[0570](0570-test-the-terminal-screens-in-a-real-terminal.md). It records four
choices made while building the Preparation tab that those records leave open.

## Context

0520 moves guided preparation into the onboarder's screens and keeps the save
rules of 0340 and 0350: stage several changes, review them, save once, and write
nothing before that. Building the tab met four cases that no record settles.
An agent made each choice while the person who owns the decisions was away. They
are proposals until that person accepts them.

## Decision

**The reviewed save is the consent to create a person's folder.** When a person
has no preparation file yet, saving creates their folder. The review says so in
words: "Saving creates the folder for this person." The key that saves, `y`, is
the consent. The store contract allows a separate consent step, and this does not
add one, because the review already names the effect before the key.

**A stale save reloads, replays, and asks again.** The plan walk saves the plan
after every answer. A person can leave the Preparation tab, answer a question,
and come back with changes staged, and the store then reports that the files
changed. The screen reloads the person from disk, applies the staged changes again
through the preparation editor, and drops any that no longer apply. It shows the
review again and needs a second `y`. Without this, staged work would become
unsaveable after any visit to another tab. A save never writes bytes the person
did not review.

**Six tabs can clip a narrow header.** The onboarder's tab strip is Walk (called Answer in 0610),
Overview, Review, Sections, Contacts, and Preparation. At 60 columns it no longer
fits. The strip is cut after the last tab that fits and ends with an ellipsis. The
title "Onboarding plan" is dropped only when keeping it would hide the current tab.
The existing narrow-window scenario still passes, because the five-tab text it
waits for survives at that width.

**The Preparation tab opens with `t`.** The keys `w o r s c` open tabs, `p n d x q`
act on the question screen, `a` and `e` act on Contacts, and `y` confirms. The
letter `i` is kept for the Invitation tab of
[0610](0610-people-work-in-screens-and-flags-serve-scripts.md).

**A terminal scenario covers the tab.** 0520 expected screens to be judged by key
and golden tests, and said binary scenarios for guided preparation ended with the
text guide. 0570 later made a real terminal the way to test screens. One scenario
adds a person, edits, reviews, and saves through the built binary.

## Consequences

- A person who wants to approve folder creation separately has no way to do it.
  A later record can add a second key if that is needed.
- Staged changes survive a visit to another tab and are lost only on quit, or when
  a person confirms dropping them with escape.
- The tab strip's behavior at narrow widths now has a rule, and the terminal page
  states it.
