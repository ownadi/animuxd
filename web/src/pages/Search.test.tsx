import React from "react";
import Search from "./Search";
import { render, fireEvent } from "@testing-library/react";
import { ThemeProvider } from "emotion-theming";
import theme from "../styles/theme";

jest.mock("../services/nibl");
jest.mock("../services/animuxd");

const SEARCH_INPUT_PLACEHOLDER_TEXT = "Search...";

it("renders search input", () => {
  const { queryByPlaceholderText, queryByTestId } = render(
    <ThemeProvider theme={theme}>
      <Search />
    </ThemeProvider>
  );

  expect(queryByPlaceholderText(SEARCH_INPUT_PLACEHOLDER_TEXT)).toBeTruthy();
  expect(queryByTestId("searchResultsRow")).toBeFalsy();
});

it("renders search results", async () => {
  const { getByPlaceholderText, findAllByTestId } = render(
    <ThemeProvider theme={theme}>
      <Search />
    </ThemeProvider>
  );

  const searchInput = getByPlaceholderText(SEARCH_INPUT_PLACEHOLDER_TEXT);
  fireEvent.change(searchInput, { target: { value: "f0o" } });
  fireEvent.submit(searchInput);

  const rows = await findAllByTestId("searchResultsRow");
  expect(rows.length).toBeGreaterThan(0);
  rows.forEach((row, idx) => {
    expect(row).toHaveTextContent(new RegExp(`f0o 0${idx + 1}`));
  });
});

it("renders summary", async () => {
  const { getByPlaceholderText, findByTestId } = render(
    <ThemeProvider theme={theme}>
      <Search />
    </ThemeProvider>
  );

  const searchInput = getByPlaceholderText(SEARCH_INPUT_PLACEHOLDER_TEXT);
  fireEvent.change(searchInput, { target: { value: "f0o" } });
  fireEvent.submit(searchInput);

  const summary = await findByTestId("searchSummary");
  expect(summary).toHaveTextContent("f0o - 2 results");
});
