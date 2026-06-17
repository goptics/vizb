// Manifest produced by the embed-ui build plugin and consumed when emitting
// pkg/template/vizb-ui.gen.go. `imports` is the gzipped-chunk reference graph
// (keyed by import-map key), `roots` maps each logical chart to its renderer
// chunk key, `entryKey` is the always-present entry chunk.
export type GoChunkArtifacts = {
  chunks: Record<string, string>
  imports: Record<string, string[]>
  roots: Record<string, string>
  entryKey: string
}
