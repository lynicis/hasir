export function getCsrfToken(): string {
  const key = "hasir-csrf";
  const cookie = document.cookie.match("(^|;)\\s*" + key  + "\\s*=\\s*([^;]+)");
  return cookie?.pop() ?? "";
}
