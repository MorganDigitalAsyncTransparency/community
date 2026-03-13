# Documentation quality prompt

Use this prompt at the start of a new conversation to run a language quality pass over the discourse-observer documentation. Technical correctness must already be verified before running this pass.

---

Go through all documentation in `discourse-observer/`. Technical correctness has already been verified — focus only on **language quality**.

For each file, identify:

- Awkward or unclear sentences that can be rewritten
- Unnecessary repetition of the same phrasing or idea within a document
- Inconsistent tone — the target tone is neutral, professional, and direct (not marketing language, not overly formal)
- Sentences that say too little or too much relative to their context
- Headings or list items that do not carry their weight
- English that is not idiomatic for technical documentation

Do **not**:

- Change technical content or decisions
- Change code examples, file names, or system names
- Add information that is not already present

Scope: `discourse-observer/` only (not the repository root).

Work file by file. Propose changes with reasoning — implement after receiving approval per file.
