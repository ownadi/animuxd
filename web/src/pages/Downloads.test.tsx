import React from "react";
import Downloads from "./Downloads";
import { render, fireEvent } from "@testing-library/react";
import { ThemeProvider } from "emotion-theming";
import theme from "../styles/theme";
import { getDownloads } from "../services/animuxd";
import { Download } from "../services/animuxdData";

jest.mock("../services/nibl");
jest.mock("../services/animuxd");

beforeAll(() => {
  jest.useFakeTimers();
});

afterEach(() => {
  jest.clearAllTimers();
});

afterAll(() => {
  jest.useRealTimers();
});

it("renders downloads", async () => {
  const { queryByPlaceholderText, findAllByTestId } = render(
    <ThemeProvider theme={theme}>
      <Downloads />
    </ThemeProvider>
  );

  const rows = await findAllByTestId("downloadsRow");
  expect(rows.length).toBe(2);

  // Depends on mock's data
  expect(rows[0]).toHaveTextContent(/foo.mkv/);
  expect(rows[0]).toHaveTextContent(/Downloading/);
  expect(rows[0]).toHaveTextContent(/50\.00%/);
  expect(rows[0]).toHaveTextContent(/3 MB\/s/);

  expect(rows[1]).toHaveTextContent(/bar.mkv/);
  expect(rows[1]).toHaveTextContent(/Waiting/);
  expect(rows[1]).toHaveTextContent(/2 GB/);
});

it("refreshes downloads every second", () => {
  render(
    <ThemeProvider theme={theme}>
      <Downloads />
    </ThemeProvider>
  );

  const initFetchedTimes = (getDownloads as jest.Mock<Promise<Download[]>, []>)
    .mock.calls.length;

  jest.advanceTimersByTime(999);
  expect(getDownloads).toHaveBeenCalledTimes(initFetchedTimes);

  jest.advanceTimersByTime(1);
  expect(getDownloads).toHaveBeenCalledTimes(initFetchedTimes + 1);
});
