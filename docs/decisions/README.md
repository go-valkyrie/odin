# Architecture Decisions

This directory contains Architecture Decision Records (ADRs) for Odin. ADRs
capture the reasoning behind significant architectural choices so that future
contributors can understand *why* the code looks the way it does, not just
*what* it does.

## When to write an ADR

Write an ADR when a decision is:

- Hard to reverse (changes a public interface, file format, or data contract)
- Shapes how contributors reason about extensions
- Has rationale that won't be obvious from reading the code
- Rejects plausible-sounding alternatives for non-obvious reasons

Small, localized, or easily-reversible decisions do not need an ADR. Commit
messages and code comments are the right home for those.

## Conventions

- Files are named `NNNN-short-title.md` with zero-padded sequential numbering.
- Each ADR has a **Status**: Proposed, Accepted, Superseded, or Rejected.
- **Never edit a decided ADR's Context, Decision, or Consequences sections.**
  If a decision changes, write a new ADR that supersedes the old one, and
  update only the status line of the superseded ADR to point at the
  superseder. This keeps the history honest.
- A Proposed ADR is a working document and may be edited freely until its
  status is changed to Accepted or Rejected.
- When a set of ADRs is tightly related (e.g. all touch the same schema
  generation), an **umbrella ADR** captures the cohort relationship and
  shared rationale. Constituent ADRs reference the umbrella; the
  umbrella is also surfaced in the by-cohort index below.

## Index — chronological

| #    | Title                                                       | Status    |
| ---- | ----------------------------------------------------------- | --------- |
| 0001 | [Bundle composition](0001-bundle-composition.md)            | Proposed  |
| 0002 | [Multi-instance support](0002-multi-instance-support.md)    | Proposed  |
| 0003 | [Namespace as a schema field](0003-namespace-as-schema-field.md) | Proposed  |
| 0004 | [Standard recommended labels](0004-standard-recommended-labels.md) | Proposed  |
| 0005 | [Bundle defaults layering](0005-bundle-defaults-layering.md)       | Proposed  |
| 0006 | [`v1alpha2` schema cohort](0006-v1alpha2-cohort.md)         | Proposed  |

## Index — by cohort

### `v1alpha2` schema cohort — [0006](0006-v1alpha2-cohort.md)

The cumulative schema evolution shipping as `api/v1alpha2`. Composition
forces these to land together; the GVK is the cross-version gate.

- 0001 — Bundle composition (cohort trigger)
- 0002 — Multi-instance support
- 0003 — Namespace as a schema field
- 0004 — Standard recommended labels
- 0005 — Bundle defaults layering
