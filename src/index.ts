import http from "node:http"
import express from "express"
import helmet from "helmet"
import cors from "cors"
import cookieSession from "cookie-session"
import { WebSocketServer } from "ws"
import dotenv from "dotenv"
import msgpack from "express-msgpack"
import { pack, unpack } from "msgpackr"

import authRoutes from "./routes/authRoutes.ts"
import appRoutes from "./routes/appRoutes.ts"
import * as initializers from "./initializers.ts"
import { RFSController } from "./controllers/app/rfsController.ts"

if (process.env.NODE_ENV !== "remote_test") {
  dotenv.config({
    path: process.env.NODE_ENV === "test" ? ".env.test" : ".env",
    quiet: true,
  })
}

await initializers.InitApp()

const app = express()

app.use(express.json())

app.use(
  helmet({
    crossOriginResourcePolicy: {
      // policy: "cross-origin", /* for production */
    },
  })
)

app.use(
  cors({
    // origin:     "http://localhost:5173", /* production client host */
    // credentials: true
  })
)

app.use(
  cookieSession({
    secret: process.env.COOKIE_SECRET,
    secure: process.env.NODE_ENV === "production",
    httpOnly: true,
  })
)

app.use("/api/auth", authRoutes)

app.use("/api/app", appRoutes)

// Create the HTTP server manually
const server = http.createServer(app)

// Attach WebSocket server to the same HTTP server
const wss = new WebSocketServer({ server, path: "/ws" })

wss.on("connection", RFSController)

let PORT: number

if (process.env.NODE_ENV != "production") {
  PORT = 8000
} else {
  PORT = parseInt(process.env.PORT || "0")
}

// Start listening
if (process.env.NODE_ENV !== "test" && process.env.NODE_ENV !== "remote_test") {
  server.listen(PORT, () => {
    console.log(`HTTP + WS server running on http://localhost:${PORT}`)
  })
}

server.on("close", () => {
  initializers.CleanUp()
})

server.on("error", () => {
  initializers.CleanUp()
})

export default server
