# i9rfs (API Server)

[![Test i9rfs](https://github.com/the-real-i9/i9rfs-server/actions/workflows/test.yml/badge.svg)](https://github.com/the-real-i9/i9rfs-server/actions/workflows/test.yml)

Build a Cloud Storage and Online File System application like Google Drive and One Drive

## Intro

i9rfs is an API server for a Remote File System application, built with entirely with WebSocket in Go. It supports major file system operations that can be used to implement Cloud Storage applications such as Google Drive and One Drive.

### Who is this project for?

If you're a frontend developer looking to build an app similar to Google Drive and One Drive, not just to have it static, but also to make it function.

The API documentation provides a detailed usage guide. It doesn't follow the Open API specification, rather, it follows Google's API documentation style, well-sturcured in a markdown table&#x2014;which I consider easier to work with.

### Open to suggestions

If you need a feature not currently supported, feel free to suggest them. It will be added as soon as possible.

## Technologies

<div style="display: flex;">
<img style="margin-right: 10px" alt="go" width="50" src="./z_attachments/tech-icons/go-original-wordmark.svg" />
<img style="margin-right: 10px" alt="go" width="50" src="./z_attachments/tech-icons/gofiber.svg" />
<img style="margin-right: 10px" alt="go" width="50" src="./z_attachments/tech-icons/websocket.svg" />
<img style="margin-right: 10px" alt="neo4j" width="100" src="./z_attachments/tech-icons/neo4j-original-wordmark.svg" />
<img style="margin-right: 10px" alt="go" width="50" src="./z_attachments/tech-icons/jwt.svg" />
<img style="margin-right: 10px" alt="go" width="50" src="./z_attachments/tech-icons/googlecloud-original.svg" />
</div>

## Table of Contents

- [Intro](#intro)
- [Technologies](#technologies)
- [Table of Contents](#table-of-contents)
- [Features](#features)
- [API Documentation](API%20doc.md)
- [Notable Features and their Algorithms](#notable-features-and-their-algorithms)
- [Building & Running the Application (Locally)](#building--running-the-application-locally)
- [Deploying the Application](#deploying-the-application)

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

## Notable Features and their Algorithms

## Building & Running the Application (Locally)

## Deploying the Application
