import { describe, it, expect } from "vitest";
import {
  AREA_KEYS,
  getGuides,
  getCategories,
  getGuide,
  searchGuides,
  MODULE_GUIDE,
  getGuideSlugForRoute,
} from "./index.js";

const VALID_BLOCK_TYPES = new Set(["p", "h2", "ul", "ol", "callout"]);
const VALID_CALLOUT_VARIANTS = new Set(["info", "tip", "warning"]);

describe("getGuides", () => {
  it("returns guides for each area sorted by category then order", () => {
    for (const area of AREA_KEYS) {
      const guides = getGuides(area);
      expect(guides.length).toBeGreaterThanOrEqual(3);
      for (let i = 1; i < guides.length; i++) {
        const prev = guides[i - 1];
        const cur = guides[i];
        const byCategory = prev.category.localeCompare(cur.category, "id");
        expect(byCategory).toBeLessThanOrEqual(0);
        if (byCategory === 0) {
          expect((prev.order ?? 0) <= (cur.order ?? 0)).toBe(true);
        }
      }
    }
  });

  it("returns an empty array for an unknown area", () => {
    // @ts-expect-error — deliberately invalid area
    expect(getGuides("nope")).toEqual([]);
  });
});

describe("getCategories", () => {
  it("groups an area's guides into at least two categories", () => {
    for (const area of AREA_KEYS) {
      const grouped = getCategories(area);
      expect(Object.keys(grouped).length).toBeGreaterThanOrEqual(2);
      const total = Object.values(grouped).reduce((n, list) => n + list.length, 0);
      expect(total).toBe(getGuides(area).length);
    }
  });
});

describe("getGuide", () => {
  it("finds a guide by slug within its area", () => {
    const guide = getGuide("app", "mengelola-data-jamaah");
    expect(guide).toBeDefined();
    expect(guide?.title).toContain("Jamaah");
  });

  it("returns undefined for a missing slug", () => {
    expect(getGuide("app", "slug-yang-tidak-ada")).toBeUndefined();
  });
});

describe("area isolation", () => {
  it("never leaks one area's guide into another", () => {
    const appSlug = "mengelola-data-jamaah";
    expect(getGuide("app", appSlug)).toBeDefined();
    // The same slug must NOT resolve from a different area.
    expect(getGuide("portal", appSlug)).toBeUndefined();
    expect(getGuide("agency", appSlug)).toBeUndefined();
  });

  it("keeps slug sets disjoint across areas", () => {
    const sets = AREA_KEYS.map((a) => new Set(getGuides(a).map((g) => g.slug)));
    for (let i = 0; i < sets.length; i++) {
      for (let j = i + 1; j < sets.length; j++) {
        for (const slug of sets[i]) {
          expect(sets[j].has(slug)).toBe(false);
        }
      }
    }
  });

  it("only searches within the requested area", () => {
    // "jamaah" exists in app content; searching portal must not return app guides.
    const appHits = searchGuides("app", "jamaah");
    expect(appHits.length).toBeGreaterThan(0);
    const portalHits = searchGuides("portal", "jamaah");
    for (const g of portalHits) {
      expect(getGuide("portal", g.slug)).toBeDefined();
    }
  });
});

describe("searchGuides", () => {
  it("returns the full sorted list for an empty query", () => {
    expect(searchGuides("app", "   ")).toEqual(getGuides("app"));
  });

  it("matches case-insensitively across fields", () => {
    const hits = searchGuides("app", "INVOICE");
    expect(hits.some((g) => g.slug === "invoice-dan-pembayaran")).toBe(true);
  });

  it("ranks a title/keyword hit above a body-only hit", () => {
    // "rooming" is in the manifest guide's title; it should outrank any guide
    // that only mentions it in the body.
    const hits = searchGuides("app", "rooming");
    expect(hits[0].slug).toBe("menyusun-manifest-rooming");
  });

  it("requires every term to match somewhere (AND semantics)", () => {
    const hits = searchGuides("app", "jamaah inisangatlangkasekali");
    expect(hits).toEqual([]);
  });

  it("returns nothing for a term that appears nowhere", () => {
    expect(searchGuides("portal", "qwertyuiop")).toEqual([]);
  });
});

describe("content integrity", () => {
  it("every guide has the required fields and valid blocks", () => {
    for (const area of AREA_KEYS) {
      for (const g of getGuides(area)) {
        expect(typeof g.slug).toBe("string");
        expect(g.slug).toMatch(/^[a-z0-9-]+$/);
        expect(g.title.length).toBeGreaterThan(0);
        expect(g.category.length).toBeGreaterThan(0);
        expect(g.summary.length).toBeGreaterThan(0);
        expect(Array.isArray(g.keywords)).toBe(true);
        expect(Array.isArray(g.body)).toBe(true);
        expect(g.body.length).toBeGreaterThan(0);
        for (const block of g.body) {
          expect(VALID_BLOCK_TYPES.has(block.type)).toBe(true);
          if (block.type === "ul" || block.type === "ol") {
            expect(Array.isArray(block.items)).toBe(true);
            expect(block.items.length).toBeGreaterThan(0);
          } else if (block.type === "callout") {
            expect(VALID_CALLOUT_VARIANTS.has(block.variant)).toBe(true);
            expect((block.text ?? "").length).toBeGreaterThan(0);
          } else {
            expect((block.text ?? "").length).toBeGreaterThan(0);
          }
        }
      }
    }
  });

  it("has unique slugs within each area", () => {
    for (const area of AREA_KEYS) {
      const slugs = getGuides(area).map((g) => g.slug);
      expect(new Set(slugs).size).toBe(slugs.length);
    }
  });

  it("every related slug resolves within the same area", () => {
    for (const area of AREA_KEYS) {
      for (const g of getGuides(area)) {
        for (const slug of g.related ?? []) {
          expect(getGuide(area, slug)).toBeDefined();
        }
      }
    }
  });
});

describe("module route mapping", () => {
  it("maps every route segment to a real guide in the same area", () => {
    for (const area of AREA_KEYS) {
      const map = MODULE_GUIDE[area] ?? {};
      for (const [segment, slug] of Object.entries(map)) {
        expect(getGuide(area, slug), `${area}/"${segment}" -> ${slug}`).toBeDefined();
      }
    }
  });

  it("resolves known segments and returns undefined otherwise", () => {
    expect(getGuideSlugForRoute("app", "")).toBe("mengenal-dashboard");
    expect(getGuideSlugForRoute("app", "jamaah")).toBe("mengelola-data-jamaah");
    expect(getGuideSlugForRoute("portal", "dokumen")).toBe("mengunggah-dokumen");
    expect(getGuideSlugForRoute("agency", "komisi")).toBe("melihat-komisi");
    expect(getGuideSlugForRoute("app", "tidak-ada")).toBeUndefined();
    expect(getGuideSlugForRoute("app", "bantuan")).toBeUndefined();
  });
});
