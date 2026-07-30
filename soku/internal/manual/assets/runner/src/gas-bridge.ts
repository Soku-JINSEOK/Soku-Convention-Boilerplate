export type GasFixture = {
  methods: Record<
    string,
    {
      result?: unknown;
      error?: string;
      mutate?: Record<string, unknown>;
    }
  >;
  state?: Record<string, unknown>;
};

export function gasBridgeSource(fixture: GasFixture): string {
  const serialized = JSON.stringify(fixture).replaceAll("<", "\\u003c");
  return `(() => {
    const fixture = ${serialized};
    const state = Object.assign({}, fixture.state || {});
    const makeRunner = (handlers = {}) => new Proxy({}, {
      get(_target, property) {
        if (property === "withSuccessHandler") return (fn) => makeRunner({...handlers, success: fn});
        if (property === "withFailureHandler") return (fn) => makeRunner({...handlers, failure: fn});
        if (property === "withUserObject") return (value) => makeRunner({...handlers, userObject: value});
        return (...args) => {
          queueMicrotask(() => {
            const operation = fixture.methods[String(property)];
            if (!operation) {
              handlers.failure?.(new Error("Unregistered GAS method: " + String(property)), handlers.userObject);
              return;
            }
            if (operation.mutate) Object.assign(state, operation.mutate);
            if (operation.error) {
              handlers.failure?.(new Error(operation.error), handlers.userObject);
              return;
            }
            const result = operation.result === "$state" ? structuredClone(state) : structuredClone(operation.result);
            handlers.success?.(result, handlers.userObject);
          });
        };
      }
    });
    globalThis.google = globalThis.google || {};
    globalThis.google.script = globalThis.google.script || {};
    globalThis.google.script.run = makeRunner();
    globalThis.__sokuGasState = state;
  })();`;
}
