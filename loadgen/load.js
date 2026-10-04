import http from "k6/http";
import { check, sleep } from "k6";

function integer(name, fallback, minimum) {
  const value = Number(__ENV[name] || fallback);
  if (!Number.isInteger(value) || value < minimum) throw new Error(`${name} must be an integer >= ${minimum}`);
  return value;
}
export const options = {
  vus: integer("VUS", 5, 1), duration: __ENV.DURATION || "10m",
  thresholds: { checks: ["rate==1"] },
};
const targets = (__ENV.TARGETS || "http://go127-v1:8080,http://go127-v1-nojsonv2:8080,http://go127-v2:8080")
  .split(",").map((url) => url.trim().replace(/\/$/, ""));
if (targets.some((url) => !/^https?:\/\//.test(url))) throw new Error("TARGETS must contain HTTP base URLs");
const profiles = (__ENV.PAYLOADS || "small,medium,large,unicode,nested").split(",").map((s) => s.trim());
const counts = { small: 1, medium: 100, large: 1000, unicode: 100, nested: 100 };
const overrideCount = __ENV.ITEMS === undefined ? null : integer("ITEMS", 100, 0);
const textSize = integer("TEXT_SIZE", 64, 0);
const pause = Number(__ENV.SLEEP || 0);
if (!Number.isFinite(pause) || pause < 0) throw new Error("SLEEP must be >= 0");

// Serialize fixtures once, outside the request loop.
const payloads = profiles.map((profile) => {
  if (!Object.prototype.hasOwnProperty.call(counts, profile)) throw new Error(`Unknown profile: ${profile}`);
  const count = overrideCount === null ? counts[profile] : overrideCount;
  const unicode = profile === "unicode";
  const body = JSON.stringify({
    query: unicode ? 'Поиск 日本語 🚀 "цитата"\nстрока' : "test query",
    items: Array.from({ length: count }, (_, i) => ({
      id: i, name: unicode ? `Товар 日本語 🚀 ${i}` : `item-${i}`, price: i * 10.25,
      tags: unicode ? ["новинка", "日本語", "🚀"] : ["one", "two", "three"],
      attributes: {
        category: "demo", active: i % 2 === 0, score: i / 10,
        description: (unicode ? "я" : "x").repeat(textSize),
        ...(profile === "nested" ? {
          details: { stock: { available: true, warehouses: [1, 2, 3] },
            variants: [{ color: "red", sizes: ["S", "M", "L"] }] },
        } : {}),
      },
    })),
  });
  return { profile, count, body };
});

export default function () {
  const payload = payloads[(__ITER + __VU - 1) % payloads.length];
  // Same JSON for every service; rotate order between iterations.
  for (let i = 0; i < targets.length; i++) {
    const target = targets[(__ITER + i) % targets.length];
    const tags = { service: target, payload: payload.profile };
    const res = http.post(`${target}/api/process`, payload.body, {
      headers: { "Content-Type": "application/json" }, tags,
    });
    check(res, {
      "status is 200": (r) => r.status === 200,
      "item count matches": (r) => {
        try { return r.json("item_count") === payload.count; } catch (_) { return false; }
      },
    }, tags);
  }
  if (pause > 0) sleep(pause);
}
