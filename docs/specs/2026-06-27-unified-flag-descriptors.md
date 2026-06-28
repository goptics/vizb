# Spec: one flag-descriptor system for all CLI flags

> Status: implementation contract. NOT committed (per CLAUDE.md).
> Branch: `refactor/chart-flag-registry` (extends the uncommitted Phase-1 work).

## Goal

Every vizb CLI flag — chart options AND parser/grouping/metadata options — is
declared once as a `Flag` descriptor and registered explicitly per command.
No "universal" bucket, no hand-written `CommonOptions/LinearOptions/ChartOptions`
`Bind`+`validationRules`. One binder, one validator, one read path.

## Layering (hard constraint)

`shared` imports `config/charts`. So the generic `Flag` type cannot live in
`shared`. It lives in **new leaf pkg `config/flags`** (stdlib-only). Both
`config/charts` (chart descriptors) and `cmd/cli` (data descriptors) import it.
No cycle: `config/flags` imports nothing internal.

Parser-specific validators (`parser.ValidateGroupPattern`, `validateParser`) and
`shared.ValidStatMath` are referenced only at descriptor *definition* sites in
`cmd/cli` — never inside `config/flags` or `config/charts`.

## The descriptor — `config/flags`

```go
type Kind int
const ( KindString Kind = iota; KindBool; KindFloat; KindStringSlice; KindStat )

type Flag struct {
    Name, Shorthand, Usage string
    Kind    Kind
    Default any                  // bind default; also warn-and-default fallback
    // chart-payload application: set ⇒ flag contributes to the chart seed
    JSONKey string
    Encode  func(any) any        // typed value → JSON-primitive payload; nil = identity
    // fatal validation (chart flags): invalid ⇒ ExitWithError
    Validate func(string) error
    // warn-and-default validation (data flags): invalid ⇒ warn + reset to Default
    ValidSet     []string
    Normalizer   func(string) string
    SoftValidate func(string) error
}

func (f Flag) IsChart() bool { return f.JSONKey != "" }
func (f Flag) IsSoft()  bool { return f.ValidSet != nil || f.Normalizer != nil || f.SoftValidate != nil }
```

Invariants:
- A descriptor has `JSONKey` set (chart flag → seed) XOR is a data flag
  (no `JSONKey`, read back by name to build `parser.Config`/metadata).
- Fatal `Validate` and the soft trio are mutually exclusive per flag.
- `KindStat` is the optional-value `--stat`: bind with `NoOptDefVal="all"`.

## FlagBag — `cmd/cli` (replaces flagbag.go + option structs)

Holds one typed pointer per flag, keyed by name:

```go
type FlagBag struct { flags []flags.Flag; strs map[string]*string;
    bools map[string]*bool; floats map[string]*float64; slices map[string]*[]string }
func NewFlagBag(fl []flags.Flag) *FlagBag
func (b *FlagBag) Bind(fs *pflag.FlagSet)            // by Kind; stat → statValue + NoOptDefVal
func (b *FlagBag) Validate()                         // fatal → ExitWithError; soft → utils.ApplyValidationRules
func (b *FlagBag) String(name) string
func (b *FlagBag) StringSlice(name) []string
func (b *FlagBag) Bool(name) bool
func (b *FlagBag) Changed(cmd, name) bool
func (b *FlagBag) ChartSeed(cmd *cobra.Command) map[string]any
func (b *FlagBag) ParseConfig() parser.Config        // = old CommonOptions.ParseConfig, sourced from bag
```

### ChartSeed rules (per chart flag with JSONKey)

- Emit `Encode(value)` into the seed map when the flag is `Changed`, OR when it
  carries a non-nil `Default` (so e.g. `scale` defaults to `"linear"` exactly as
  today via the descriptor `Default`, not a special case).
- `KindStat` tri-state (matches `MaterialiseStatFlags`, emitted as a plain map):
  - not changed ⇒ omit `stat`
  - `--stat` (alone) ⇒ `{"enabled":true,"math":[]}`
  - `--stat=a,b` ⇒ `{"enabled":true,"math":["a","b"]}`

### Validate rules

- Fatal flag: if `Validate(value)!=nil` → `shared.ExitWithError`.
  Float flags (symbol-size) validate the formatted number.
- Soft flag: build a `utils.ValidationRule{Value/SliceValue: bag ptr, ValidSet,
  Validator: SoftValidate, Normalizer, Default/SliceDefault: Default}` and run
  `utils.ApplyValidationRules` (warn + mutate pointer). Identical UX to today.
- Scale becomes a soft chart flag (`ValidSet:[linear log]`, `Normalizer:ToLower`,
  `Default:"linear"`) — replaces the bespoke `ValidateScale`. It still also has
  `JSONKey:"scale"` + `Encode:ToLower`, so it both warn-defaults AND seeds.

### DUAL validation — preserve the existing asymmetry (scale / sort / stat)

These three flags behave DIFFERENTLY depending on the path, and that asymmetry
is current behaviour that must be preserved:

| flag | subcommand path (`--scale`) | override path (`--chart t:scale=x`) |
|------|------------------------------|--------------------------------------|
| scale | warn-and-default → `linear` | **fatal** error |
| sort  | warn-and-default → unset    | **fatal** error |
| stat  | warn-and-default (drop bad categories) | **fatal** error |

Implementation: each of these descriptors carries BOTH the soft trio AND a fatal
`Validate`:
- `ScaleFlag`: `Validate: ValidateScaleValue` + `ValidSet/Normalizer/Default` soft.
- `SortFlag`:  `Validate: ValidateSortValue`  + `ValidSet/Normalizer` soft.
- `StatFlag`:  category validation against `shared.ValidStatMath` both ways.

Routing rule:
- **FlagBag.Validate** (subcommand path) → if `f.IsSoft()` use the soft rule
  (`utils.ApplyValidationRules`); never call fatal `Validate`. So scale/sort/stat
  warn-and-default at the subcommand.
- **ParseOverrides.convertFlagValue** (override path) → keep calling `f.Validate`
  fatally when set (current behaviour); swap stays axis-validated.

`labels` (bool) has no validation on either path.

## Descriptor catalogs

### `config/charts/flag.go` — chart descriptors (all carry JSONKey)

- Move swap/sort/labels here as before; **add `StatFlag`** (`KindStat`,
  `JSONKey:"stat"`). swap: no `Validate` (axis check done by override parser).
  sort: soft (`ValidSet:[asc desc]`, `Normalizer:ToLower`) + `Encode` →
  `{"enabled":true,"order":<lower>}`. labels: `KindBool`, `JSONKey:"showLabels"`.
- Keep `ScaleFlag/ThreeDFlag/ThreeDRotateFlag/ThreeDVisualMapFlag/VisualMapFlag/SymbolFlag/SymbolSizeFlag`.
- `var BaseChartFlags = []flags.Flag{SwapFlag, SortFlag, LabelsFlag, StatFlag}`.
- **Delete `UniversalFlags`**; `FlagsFor(type)` returns `spec.Flags` verbatim;
  `AllFlagNames` iterates every spec's flags.

### `config/charts/<c>/<c>.go` (×6)

`Flags: append(slices.Clone(charts.BaseChartFlags), <variable flags>)`
- bar/line: `+ ScaleFlag, ThreeDFlag, ThreeDRotateFlag, ThreeDVisualMapFlag`
  (line also `+ SymbolFlag, SymbolSizeFlag`)
- scatter: `+ VisualMapFlag, SymbolFlag, SymbolSizeFlag`
- pie/heatmap/radar: base only.

### `cmd/cli/dataflags.go` (new) — data/metadata descriptors

`flags.Flag` values, no JSONKey:
- meta: `name(-n,def "Comparisons"), description(-d), output(-o), tag(-t)`
- parser: `parser(-P,"auto", SoftValidate:validateParser, Default:"auto")`
- grouping: `group-pattern(-p,"x", SoftValidate:parser.ValidateGroupPattern, Default:"xAxis")`,
  `group-regex(-r)`, `group(-g, KindStringSlice)`, `filter(-f)`
- units: `mem-unit(-M,"B", ValidSet:[b B KB MB GB], Normalizer:kb→KB…, Default:"B")`,
  `time-unit(-T,"ns", ValidSet:[ns us ms s], Default:"ns")`,
  `number-unit(-N, ValidSet:[K M B T], Normalizer:ToUpper, Default:"")`
- selection: `select`, `json-path`
Helper slices `DataFlags` (all of the above).

## Consumers

- `shared/chart_spec.go`: type alias → `flags.Flag`; behaviour unchanged
  (drop-with-warning). `stat`/`stat=a,b` now valid `--chart` keys (encode via
  KindStat). swap still axis-validated here.
- `cmd/cli/command.go`: `bag := NewFlagBag(append(slices.Clone(DataFlags), spec.Flags...))`;
  Run → `bag.Validate()`, `seed := bag.ChartSeed(cmd)`, `Materialise(spec.Type, seed, nil)`,
  swap-vs-axes, `RunSingleChart(cmd,args,meta,cfg)` where `meta` = small carrier
  built from bag (Name/Description/Tag/Output/Parser). Delete `universalSeed`,
  `ValidateScale`.
- `cmd/root.go`: `rootBag := NewFlagBag(DataFlags + [SortFlag,LabelsFlag,StatFlag] + charts/chart descriptors)`;
  seed via `rootBag.ChartSeed`; loop `Materialise`. Keep root sort/labels deprecation warnings.
- `cmd/cli/pipeline.go`, `cmd/ui.go`: swap `CommonOptions` carrier for the small
  meta struct + `bag.ParseConfig()`; pipeline logic unchanged.

## Behaviour parity (must hold — golden diff)

Byte-identical `settings`/`axes`/`name` JSON for all VALID invocations:
- `bar --scale=log --3d --sort=asc --stat` → `[{type:bar,sort:{enabled,order:asc},scale:log,threeD:true,stat:{enabled:true,math:[]}}]`
- `scatter --symbol=diamond --symbol-size=12`
- `--charts bar,scatter --chart bar:scale=log,sort=asc --chart scatter:symbol=diamond`
- `-M KB -T us -n MyCmp` and csv `-p 'x,n'` (axes/units unchanged)
- warn-and-default: `-M bogus` → stderr warn + `B`, exit 0
- drop-with-warning: `--chart pie:scale=log` → warn+ignore; `--chart bar:bogus` → error

Accepted message-only change: `--chart bar:bogus` valid-key list now also lists
`stat` (stat became a real chart descriptor). Value outputs unchanged.

## Out of scope

- UI / gen.go (no UI change).
- merge command flags (no chart/parser flag overlap to unify).
