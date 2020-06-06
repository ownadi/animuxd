import { atom } from "recoil";

export enum Page {
  Downloads,
  Search,
}

export const page = atom<Page>({ key: "page", default: Page.Search });
