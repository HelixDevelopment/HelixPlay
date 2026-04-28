# HelixPlay Gaming System

Slogan: "Ultimate gaming experience!"

Root directory containing documentation is docs, under it we have research materials: research.
Research is divided in phases and we are starting work on our MVP, so all materials we have prepared for this task are under:
docs/research/chapters/MVP

We have performed several researches for the HelixPlay project and all results (with parent requests) are located under:
01_base, 02_latency, and 03_video_technology directories.

Dive deep into all materials we have obtained and created - all documentation, created source code, 
diagrams, graphs, schemas and other relevant materials.

We MUST create final in-depth step by step implementation documentation with all phases of development,
all tasks and sub-tasks (fine grained), detailed steps and as much as possible detail.

We MUST merge all gathered knowledge into one big ulitmate project documentation with full specs redy for 
full development by our engineering team!

Nothing from materials we have nprovided can be omitted! Nothing can be skipped or ignored!
Simplification or bluffing is strictly forbidden!

We cannot have less material in ultimate documentation than in all 3 source research phases (directories)!

We MUST dive deeper regarding each point and context to validate, verify and extend with additional details!
We MUST use as much as possible comprehensive web research - for technical articles, howtos, research papers, opensourced codebase and
all this MUST be wisely used and properly incorporated into the project!

Once full and final documentation with all materials we have mentioned is created, validated and verified in multiple-passes we can implement the whole System!
It is not allowed to leave TODO / FIXME placholders, no dummy / placeholder classes, everything MUST be fully wired and no dead code or anything hanging left!
No skipping is allowed or bluffing of any kind! These are all MANDATORY constraints which MUST be part of the Constitution, CLAUDE.MD and AGENTS.MD!

Mandatory constraints / rules to follow:

- Everything MUST BE fully decoupled and reusable on most generic level by any other project in the future
- Decoupled components MUST BE separated in proper (or belong to) Submodules (Git / Go lang Submodules / Modules) (under vasic-digital organization on GitHub and GitLab)
- We MUST use Lazy initialization over eager wherever is reasonable and possible
- All Submdoules we create MUST be public
- Do not repeat yourself
- We are aiming to heavy use of non-blocking concurrency, proper and safe use of Semaphores and other mechanisms which will prevent our systems from clogging
- We MUST use as much as possible events and observability so all changes in the System are applied in real-time for interested parties (events being received)
- We prefer gRPC, however any other advanced Systems and technologies are welcome to be incorporated (RabbitMQ, Redis, and so on)
- HTTP3 (Quic / Cronet)
- Brotli compression
- REST API and additional middleware - separated into Micorservices which are fully decoupled
- Services discovery in a same network
- Dynamic ports assignment to Containers
- Every codebase is executed in Containers - every single Service, Infrastructure part (Database and others), Building, testing and devugging, scanning
- Heavy scanning and checks: SonarQube, Snyk and other advanced code quality and security scanning solutions
- Local only CI / CD (run inside the Containers)
- All Comntainers work handled by our Containers submodule: https://github.com/vasic-digital/Containers
- We use all available Submodules already created under vasic-digital organization: https://github.com/vasic-digital/ so we do not create Submodules (Modules) for same purposes multipel times
- If some Submodule does not have all features we need, we MUST extend it properly with additional features / functionalities (since we have full control of all vasic-digital and HelixDevelopment organizations)

Testing:

Every single Submodule and piece of code MUST BE covered 100% with the following types of the tests:

- Unit
- Integration
- E2E
- Security
- Benchmarking
- Chaos
- Stress
- Smoke
- Full automation
- Challenges, which are special type of the tests / testing which require production Containers booted up and the whole System binnaries. Challenges incorporate our Challenges Submodule:
git@github.com:vasic-digital/Challenges.git

See how Challenges have been incorporated under same Projects project's root dir for HelixAgent and Catalogizer projects!

We MUST incorporate and make work properly full autonomous QA System HelixQA and fully integrate it into the System:
git@github.com:HelixDevelopment/HelixQA.git

See how Challenges have been incorporated under same Projects project's root dir for HelixAgent and Catalogizer projects!

IMPORTANT: For every single Submodule we add we MUST add as well all its dependency Submodules! FOr every single Submodule there is comprehensive documentation and fully accessible codebase 
we can in-depth learn and analyse for easier and better incorporation!

IMPORTANT: Only Unit tests are allowed to have Mocks, Stubs, placeholder classes or hardcoded values! All other testing types and Systems MUST USE real production ready implementation and have 
the whole system (all Containers) ready, fully operational and running!

IMPORTANT: Make sure that all existing tests and Challenges do work in anti-bluff manner - they MUST confirm that all tested codebase really works as expected! 
We had been in position that all tests do execute with success and all Challenges as well, but in reality the most of the features does not work and can't be used! 
This MUST NOT be the case and execution of tests and Challenges MUST guarantee the quality, the completition and full usability by end users of the product! 
This MUST BE part of Constitution of our project, its CLAUDE.MD and AGENTS.MD if it is not there already, and to be applied to all Submodules's Constitutuon, CLAUDE.MD and AGENTS.MD as well (if not there already)!

Implementation steps:

All work MUST be performed by carefully planned phases divided into fine grained tasks and sutasks. Not a single detail can be omitted from the specifications and task definitions.
All implementation we get MUST be validated and verified to the smallest details against the project documentation, specifications and other materials we have created for this project! No skipping or bluffing in this matter is allowed!
Since we have full access to GitHub and GitLab CLIs all implementation phases and tasks with sub-tasks MUST be tracked through GitHub Project and GitLab equivalent if CLIs are allowing us to create working tickets and modify them!
Final state of implementation work MUST be clean board of Project with all phases and tickets fully addressed without anything skipped, simplified, omitted, disabled!

