<!-- omit in toc -->
# Users microservice

<!-- omit in toc -->
## Table of content

- [Project structure](#project-structure)
- [Asumptions, Desitions and Things to change](#asumptions-desitions-and-things-to-change)

## Project structure

//TODO

## Asumptions, Desitions and Things to change

- User password will be clear text for simplicity: no bcrypt, no hashing, no hiding in JSON responses, etc. This is for the same reasoning the login is not provided.
- There should be more edge cases when testing
- Despite the text saying we must use "id", I used "_id". There are some workarounds that could be done but for simplicity for this challenge, I didn't do it. Some workarounds:
  - Switching to MySQL
  - Adding another field called "id", making it unique and "forgetting" about "_id" (also creating and index for search)
  - Create 2 fields (_id and id) and sync them with the same value
- There should be a "use case" layer, in which we test and execute the business logic for each use case. In this project this layer does not exists, all the logic has been done in the handlers layer. This has some side effects, such as:
  - We cannot correctly test the handler logic and the use case logic without changing the same test
  - We cannot reuse the logic in other parts of the application if needed.
  - If the use case layer has a new dependency, we have to modify the handlers instead.
