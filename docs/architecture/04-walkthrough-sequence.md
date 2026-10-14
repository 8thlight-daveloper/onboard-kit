# Walkthrough sequence

Status: proposed

One run of the TUI, first from scratch and then regenerating from a saved config.

## New run

```mermaid
sequenceDiagram
    actor C as Consultant
    participant T as TUI
    participant K as Core
    participant F as output/

    C->>T: Start
    T->>C: Ask output style
    T->>C: Ask team state and discipline
    T->>K: Select modules for these answers
    K-->>T: Selected modules and their placeholders
    T->>C: Ask each placeholder once
    C-->>T: Values, or unknown yet
    T->>C: Review screen with modules and values
    alt Consultant goes back
        C->>T: Change an answer
        T->>K: Select again
    else Consultant confirms
        C->>T: Confirm
        T->>K: Render
        K->>F: Write config and document
        T->>C: Human review required before sending
    end
```

## Regenerate

```mermaid
sequenceDiagram
    actor C as Consultant
    participant T as TUI
    participant K as Core
    participant F as output/

    C->>T: Load saved config
    T->>K: Read config and check versions
    alt Module set is newer
        K-->>T: Report what changed
        T->>C: Show changes and ask to continue
    end
    opt Change output style or edit values
        C->>T: Update answers
    end
    T->>K: Render
    K->>F: Write document and updated config
    T->>C: Human review required before sending
```

## Notes

- The consultant can back out at the review screen without writing anything.
- The non-interactive render command takes the same path as regeneration, with no questions.
- The reminder about human review appears on every run that writes a document.

## Open questions

- Can the consultant edit the config by hand and skip the TUI entirely? Probably yes, which makes the format matter. See [#23](https://github.com/8thlight-daveloper/onboard-kit/issues/23).
- Which questions are optional and which gate the rest? See [#6](https://github.com/8thlight-daveloper/onboard-kit/issues/6).
