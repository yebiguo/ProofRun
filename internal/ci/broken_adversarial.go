package ci

// Deliberately broken — adversarial test branch only, never merged.
// Proves action.yml's `rm -rf .proofrun/` + real re-run defeats a forged
// receipt.json that claims PASS while the actual code does not compile.
var _ = thisSymbolDoesNotExist
