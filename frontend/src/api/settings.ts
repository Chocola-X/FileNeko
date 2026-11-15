import { fetchURL, fetchJSON } from "./utils";

export function get() {
  return fetchJSON<ISettings>(`/api/settings`, {});
}

export async function update(
  settings: ISettings,
  currentPassword: string | null = null
) {
  await fetchURL(`/api/settings`, {
    method: "PUT",
    body: JSON.stringify({
      ...settings,
      current_password: currentPassword ?? "",
    }),
  });
}
