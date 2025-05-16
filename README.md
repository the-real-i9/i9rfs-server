# i9rfs (API Server)

[![Test i9rfs](https://github.com/the-real-i9/i9rfs-server/actions/workflows/test.yml/badge.svg)](https://github.com/the-real-i9/i9rfs-server/actions/workflows/test.yml)

Build a Cloud Storage and Online File System application like Google Drive and One Drive

## Intro

i9rfs is an API server for a Virtual File System application, built with entirely using WebSocket in Go. It supports major file system operations that can be used to implement Cloud Storage applications such as Google Drive and One Drive.

## Technologies

<div style="display: flex;">
<img style="margin-right: 10px" alt="go" width="40" src="./.attachments/tech-icons/go-original-wordmark.svg" />
<img style="margin-right: 10px" alt="go" width="40" src="./.attachments/tech-icons/gofiber.svg" />
<img style="margin-right: 10px" alt="go" width="40" src="./.attachments/tech-icons/websocket.svg" />
<img style="margin-right: 10px" alt="neo4j" width="60" src="./.attachments/tech-icons/neo4j-original.svg" />
<img style="margin-right: 10px" alt="go" width="40" src="./.attachments/tech-icons/jwt.svg" />
<img style="margin-right: 10px" alt="go" width="40" src="./.attachments/tech-icons/googlecloud-original.svg" />
</div>

## Table of Contents

- [Intro](#intro)
- [Technologies](#technologies)
- [Table of Contents](#table-of-contents)
- [Features](#features)
- [API Documentation](API%20doc.md)

## Features

Visit the API documentation for implementation guide.

- **Create** directories
- **Upload** files
- **List** the contents of a directory
- **Copy** and **Move** files/directories
- **Delete** files/directories
- **Rename** files/directories
- Move files/directories to **Trash**
- View files/directories in Trash
- Restore files/directories from Trash
- **Star** files/directories

## API Documentation

For all **REST request/response Communication**: [Click Here](./.apidoc/restapi.md)

For all **WebSocket Real-time Communication**: [Click Here](./.apidoc/websocketsapi.md)
