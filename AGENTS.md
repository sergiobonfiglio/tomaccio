# Project Principles

- Keep CLI first.
- Keep small interfaces at subsystem boundaries for future extension but wire current concrete implementations directly.

# Coding guidelines

- Minimal dependencies: prefer standard library, avoid frameworks and ORMs.
- Adopt Go conventions, idioms and best practices.
- Comment your code, especially exported functions and types, non obvious logic.
- Prefer red-green test driven development.
