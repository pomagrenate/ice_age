# Contributing

Contributions to **Iceage** are welcome.

The project is intentionally small and focused: improve the prompt, improve the developer experience, keep the behavior predictable.

> **Small focused change > big rewrite. Iceage likes simple.**

## How to Contribute

### 1. Fork the repository

Fork the repository and clone your fork locally:

```bash
git clone https://github.com/YOUR_USERNAME/iceage.git
cd iceage
```

### 2. Edit the source of truth

For changes to the main Iceage behavior, edit only:

```text
skills/iceage/SKILL.md
```

This is the **single source of truth** for the Iceage prompt.

Do not manually edit generated copies.

### 3. Test your change

Run Iceage with your updated prompt and compare its output against the previous behavior.

Focus on concrete improvements such as:

* Better information density
* Less unnecessary verbosity
* More consistent technical language
* Better handling of edge cases
* Fewer unnecessary tokens
* No loss of important technical information

Avoid changes that make Iceage shorter at the cost of correctness or clarity.

### 4. Open a Pull Request

Your PR should include a small, concrete before/after example.

Use this structure:

```markdown
## Before

> What Iceage says now.

## After

> What Iceage says with this change.

## Why

One sentence explaining why the new behavior is better.
```

Keep the example representative of the behavior you are changing.

---

## Source of Truth

The following files are automatically synchronized by CI after changes are merged:

```text
skills/iceage/SKILL.md
plugins/iceage/skills/iceage/SKILL.md
.cursor/skills/iceage/SKILL.md
iceage.skill
```

**Only edit:**

```text
skills/iceage/SKILL.md
```

Do not edit the generated copies directly.

CI will synchronize them after the PR is merged.

---

## Compress Skill

The compression skill has its own source of truth.

If you are modifying `iceage-compress`, edit:

```text
skills/compress/SKILL.md
```

This is the only file you need to change for the compress skill.

Do not manually modify generated copies.

---

## What Makes a Good Contribution?

Good Iceage changes are usually:

* Small
* Focused
* Easy to understand
* Backed by a concrete example
* Easy to evaluate
* Useful across different coding tasks

A good change should answer:

> **Does this remove unnecessary words without removing useful information?**

If yes, it is probably a good fit for Iceage.

---

## What to Avoid

Avoid large prompt rewrites without a specific problem to solve.

Avoid changes that:

* Make responses cryptic
* Remove important technical context
* Reduce security warnings
* Break code examples
* Introduce unnecessary complexity
* Change unrelated behavior
* Optimize for token count alone

**Shorter is not automatically better.**

The goal is:

```text
Less words
    +
Same meaning
    +
Same technical accuracy
```

---

## Ideas & Starter Tasks

Looking for something small to work on?

Check the issues labeled:

**[good first issue](../../issues?q=label%3A%22good+first+issue%22)**

These are a good starting point for first-time contributors.

---

## Pull Request Checklist

Before opening a PR:

* [ ] Change is focused on one improvement
* [ ] `skills/iceage/SKILL.md` is the only Iceage prompt source modified
* [ ] Before/after example is included
* [ ] One-sentence rationale is included
* [ ] Technical meaning is preserved
* [ ] No unnecessary complexity was introduced
* [ ] Generated files were not manually edited

For compress-skill changes:

* [ ] `skills/compress/SKILL.md` is the source of truth
* [ ] Generated copies were not manually edited

---

## Philosophy

Iceage is deliberately simple.

If a change can be solved with a few lines instead of a large abstraction, prefer the few lines.

> **Small focused change > big rewrite.**

Thanks for helping make AI coding agents **less verbose, more precise, and easier to work with.** 🦣
