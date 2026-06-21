import { describe, it, expect } from "vitest";
import { edgeScrollVelocity } from "./dragAutoScroll.js";

const L = 100, R = 900, EDGE = 80, MAX = 20;

describe("edgeScrollVelocity", () => {
  it("no scroll in the middle of the container", () => {
    expect(edgeScrollVelocity(500, L, R, EDGE, MAX)).toBe(0);
  });

  it("no scroll exactly at the zone boundaries", () => {
    expect(edgeScrollVelocity(L + EDGE, L, R, EDGE, MAX)).toBe(0);
    expect(edgeScrollVelocity(R - EDGE, L, R, EDGE, MAX)).toBe(0);
  });

  it("scrolls left (negative) inside the left edge zone", () => {
    expect(edgeScrollVelocity(L + 10, L, R, EDGE, MAX)).toBeLessThan(0);
  });

  it("scrolls right (positive) inside the right edge zone", () => {
    expect(edgeScrollVelocity(R - 10, L, R, EDGE, MAX)).toBeGreaterThan(0);
  });

  it("is symmetric: equal depth into each edge gives opposite, equal-magnitude velocity", () => {
    const left = edgeScrollVelocity(L + 20, L, R, EDGE, MAX);
    const right = edgeScrollVelocity(R - 20, L, R, EDGE, MAX);
    expect(left).toBeLessThan(0);
    expect(left).toBe(-right);
  });

  it("ramps: deeper into the zone scrolls faster", () => {
    const shallow = Math.abs(edgeScrollVelocity(L + 60, L, R, EDGE, MAX));
    const deep = Math.abs(edgeScrollVelocity(L + 5, L, R, EDGE, MAX));
    expect(deep).toBeGreaterThan(shallow);
  });

  it("caps at maxSpeed at and beyond either edge", () => {
    expect(edgeScrollVelocity(L, L, R, EDGE, MAX)).toBe(-MAX);
    expect(edgeScrollVelocity(L - 50, L, R, EDGE, MAX)).toBe(-MAX);
    expect(edgeScrollVelocity(R, L, R, EDGE, MAX)).toBe(MAX);
    expect(edgeScrollVelocity(R + 50, L, R, EDGE, MAX)).toBe(MAX);
  });
});