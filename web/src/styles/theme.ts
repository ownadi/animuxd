import { Theme } from "./styled";
import color from "tinycolor2";

// Inspired by https://colorhunt.co/palette/118962
const primaryColor = "#283149";
const secondaryColor = "#404b69";
const accentColor = "#00818a";
const textColor = "#dbedf3";

const theme: Theme = {
  color: {
    navigationBackground: color(secondaryColor).darken(5).toHexString(),
    navigationItemBackground: secondaryColor,
    navigationActiveItemBackground: accentColor,
    navigationTextColor: textColor,
    navigationActiveColor: color(secondaryColor).darken(5).toHexString(),
    workspaceBackground: primaryColor,
    inputBorder: accentColor,
    inputFocusBorder: color(accentColor).brighten(10).toHexString(),
    inputBackground: secondaryColor,
    inputText: textColor,
    tableHeader: color(secondaryColor).darken(5).toHexString(),
    tableRow: secondaryColor,
    tableRowHover: accentColor,
    tableText: textColor,
    scrollTrack: secondaryColor,
    scrollThumb: textColor,
    statusBackground: accentColor,
    statusText: textColor,
  },
};

export default theme;
