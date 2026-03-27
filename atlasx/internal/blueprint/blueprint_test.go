package blueprint

import "testing"

func TestDefaultBlueprint(t *testing.T) {
	product := Default()
	if product.Name != "AtlasX" {
		t.Fatalf("unexpected product name: %s", product.Name)
	}
	if product.ControlPlane != "Go" {
		t.Fatalf("unexpected control plane: %s", product.ControlPlane)
	}
	if len(product.Phases) < 4 {
		t.Fatalf("expected phased rebuild plan")
	}
}
