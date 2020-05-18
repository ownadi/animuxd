import { NIBL_API_URL, NiblBot, NiblPackage } from "./niblData";

export const search = (query: string): Promise<NiblPackage[]> => {
  const url = new URL(`${NIBL_API_URL}/search`);
  url.search = new URLSearchParams({ query }).toString();

  return fetch(url.toString())
    .then((response) => response.json())
    .then((json) => json.content as NiblPackage[]);
};

export const getBots = (): Promise<NiblBot[]> => {
  return fetch(`${NIBL_API_URL}/bots`)
    .then((response) => response.json())
    .then((json) => json.content as NiblBot[]);
};
