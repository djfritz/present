# present
Modified fork of Go's present tool

- Removed golang tracker scripts
- Removed jquery
- Removed go playground directive and related scripting

Additionally, I've added a static content generator that walks the present server and generates a static directory of all content. To use this, run the present tool with the `-static` flag:

`present -content my_content -static static_output`

