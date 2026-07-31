const sensitiveName = /^(?:key|api_?key|token|access_?token|signature|sig|secret|password)$/i;
const credentialLiteral = /(AIza[0-9A-Za-z_-]{20,}|(?:key|token|signature|secret|password)=([^&\s]+))/gi;

export function redactText(value: string): string {
  return value.replace(credentialLiteral, (match, _group, offset: number, source: string) => {
    const separator = match.indexOf("=");
    if (separator >= 0) {
      return `${match.slice(0, separator + 1)}[REDACTED]`;
    }
    return "[REDACTED]";
  });
}

export function redactURL(value: string): string {
  try {
    const parsed = new URL(value);
    parsed.username = "";
    parsed.password = "";
    for (const [name] of parsed.searchParams) {
      if (sensitiveName.test(name)) {
        parsed.searchParams.set(name, "[REDACTED]");
      }
    }
    return parsed.toString();
  } catch {
    return redactText(value);
  }
}

export function assertRedacted(value: unknown): void {
  const serialized = JSON.stringify(value);
  if (/(AIza[0-9A-Za-z_-]{20,}|(?:key|token|signature|secret|password)=[^&"\s[]+)/i.test(serialized)) {
    throw new Error("durable report contains a credential-like value");
  }
}
