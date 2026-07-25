import http from 'node:http'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const root = path.join(__dirname, 'fixtures')
const port = Number(process.env.E2E_PORT || 4173)

const mime = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.json': 'application/json',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
}

const server = http.createServer((req, res) => {
  const urlPath = decodeURIComponent((req.url || '/').split('?')[0] || '/')
  const rel = urlPath === '/' ? 'index.html' : urlPath.replace(/^\//, '')
  const file = path.normalize(path.join(root, rel))
  if (!file.startsWith(root)) {
    res.writeHead(403).end('Forbidden')
    return
  }
  fs.readFile(file, (err, data) => {
    if (err) {
      res.writeHead(404).end('Not found')
      return
    }
    const ext = path.extname(file)
    res.writeHead(200, { 'Content-Type': mime[ext] || 'application/octet-stream' })
    res.end(data)
  })
})

server.listen(port, '127.0.0.1', () => {
  console.log(`e2e fixture server http://127.0.0.1:${port}`)
})
