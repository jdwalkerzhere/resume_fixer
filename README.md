# Resume Fixer

The simple Anthropic agent that makes suggestions on how to improve your resume. 

## Current Scope
- Asks for a path to a resume (Currently this only accepts `.txt` or `.md` files, and I'd suggest absolute filepath), and responds with suggested edits following resume best practices.

## Future Scope
- Accepts a path to a resume and a job description and streams back suggested edits.
- Agent can edit the files themselves
- More defined project structure (active_resumes, archived_resumes, etc).
- Asks questions to better understand the user's experience.
- Add CLI commands
- Persist Data about the user via `resume_fixer update` CLI command, makes resume updates on the fly.
- Voice Interface for candidate Q&A

## Dependencies
- Golang
- Anthropic SDK (API Key Needs to be set in environment or .env file)
- (eventually) pandoc

## Usage
1. Fork the repository
2. Set `ANTHROPIC_API_KEY` either by `export` or writen to a `.env` file
3. Run `go mod tidy` and `go build -o resume_fixer main.go`
4. Run `./resume_fixer`

## Jank
Yes, I *know* it's jank that it'll only take `.txt` and `.md` files, Ok? This is a v0 learning project, more to come (or send a PR)
