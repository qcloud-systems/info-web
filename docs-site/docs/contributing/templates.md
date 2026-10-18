---
description: Starter templates for each kind of page in the knowledge base.
---
# Page Templates

Copy the template that matches the page you are writing. Each one follows the learning path: an **Engineering Note** records what was learned, and a **Tutorial** turns it into a guided build. A **Self-Hosting Guide** makes it something administrators run, and a **Product Guide** documents it for customers.

The source files live in `docs-site/templates/`.

=== "Engineering Note"

    ````markdown
    --8<-- "note.md"
    ````

=== "Tutorial"

    ````markdown
    --8<-- "tutorial.md"
    ````

=== "Self-Hosting Guide"

    ````markdown
    --8<-- "self-hosting-guide.md"
    ````

=== "Product Guide"

    ````markdown
    --8<-- "product-guide.md"
    ````

!!! tip "Tutorials"
    Tutorials live beside the notes they build on, in the same Engineering Notes topic folder, and carry the `Tutorial` tag. Browse them all from [Tags](../tags.md).
