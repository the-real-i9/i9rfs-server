import nodemailer from "nodemailer"

export function SendMail(email: string, subject: string, body: string) {
  if (process.env.NODE_ENV === "test") {
    return
  }

  const transporter = nodemailer.createTransport({
    service: "gmail",
    auth: {
      user: process.env.MAILING_EMAIL,
      pass: process.env.MAILING_PASSWORD,
    },
  })

  transporter
    .sendMail({
      from: process.env.MAILING_EMAIL,
      to: email,
      subject: `i9rfs - ${subject}`,
      html: body,
    })
    .catch((err) => {
      console.error("SendMail error:", err)
    })
}
