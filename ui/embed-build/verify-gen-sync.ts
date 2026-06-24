import { compareGenGoSources, readGenGo } from './gen-go-parse.ts'

const usage = 'Usage: node embed-build/verify-gen-sync.ts <committed.go> <rebuilt.go>'

const main = (): void => {
  const [committedPath, rebuiltPath] = process.argv.slice(2)
  if (!committedPath || !rebuiltPath) {
    console.error(usage)
    process.exit(1)
  }

  let committed: string
  let rebuilt: string
  try {
    committed = readGenGo(committedPath)
    rebuilt = readGenGo(rebuiltPath)
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err)
    console.error(message)
    process.exit(1)
  }

  const errors = compareGenGoSources(committed, rebuilt)

  if (errors.length > 0) {
    console.error('pkg/template/vizb-ui.gen.go is out of sync with ui sources.')
    for (const err of errors) console.error(`  - ${err}`)
    console.error('Run: cd ui && EMBED_UI=True pnpm build && git add pkg/template/vizb-ui.gen.go')
    process.exit(1)
  }

  console.info('vizb-ui.gen.go chunk graph and template are in sync')
}

main()
