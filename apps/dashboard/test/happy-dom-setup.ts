import { GlobalRegistrator } from '@happy-dom/global-registrator';

GlobalRegistrator.register();

if (typeof document !== "undefined" && !document.body) {
  const body = document.createElement("body");
  const html = document.querySelector("html");
  if (html) {
    html.appendChild(body);
  }
}
