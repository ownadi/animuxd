import React, { useRef, useCallback, useState } from "react";
import styled from "../styles/styled";
import useNiblSearchResults from "../hooks/useNiblSearchResults";
import SearchResultsTable from "./Search/SearchResultsTable";
import { FaSpinner } from "react-icons/fa";

const Container = styled.div`
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
`;

const InputWrapper = styled.div`
  position: relative;
  width: 100%;
  margin-bottom: 1rem;
`;

type InputIconWrapperProps = {
  loading: boolean;
};
const InputIconWrapper = styled.div<InputIconWrapperProps>`
  position: absolute;
  right: 1rem;
  top: 1.6rem;
  transform: translateY(-50%);
  color: ${(props) => props.theme.color.inputText};
  visibility: ${(props) => (props.loading ? "initial" : "hidden")};
`;

const SearchInput = styled.input`
  width: 100%;
  box-sizing: border-box;
  border: 0.2rem solid ${(props) => props.theme.color.inputBorder};
  border-radius: 0.4rem;
  background: ${(props) => props.theme.color.inputBackground};
  color: ${(props) => props.theme.color.inputText};
  height: 3rem;
  font-size: 1.5rem;
  padding: 0 1rem;
  outline: none;
  transition: all 0.2s ease-in-out;

  &:focus {
    border: 0.2rem solid ${(props) => props.theme.color.inputFocusBorder};
  }
`;

type SearchSummaryProps = {
  ready: boolean;
};
const SearchSummary = styled.div<SearchSummaryProps>`
  text-align: right;
  color: ${(props) => props.theme.color.tableText};
  margin-top: 0.5rem;
  font-size: 1.4rem;
  visibility: ${(props) => (props.ready ? "initial" : "hidden")};
`;

const Search = () => {
  const searchInputRef = useRef<HTMLInputElement>(null);
  const [searchPhrase, setSearchPhrase] = useState<string>("");

  const { searchResults, loading, query } = useNiblSearchResults(searchPhrase);

  const onSearchSubmit = useCallback(
    (e) => {
      e.preventDefault();

      setSearchPhrase(searchInputRef.current?.value || "");
    },
    [searchInputRef]
  );

  return (
    <Container>
      <form onSubmit={onSearchSubmit}>
        <InputWrapper>
          <SearchInput placeholder="Search..." ref={searchInputRef} />
          <InputIconWrapper loading={loading}>
            <FaSpinner className="fa-spin" />
          </InputIconWrapper>
          <SearchSummary ready={!!query} data-testid="searchSummary">
            {`${query} - ${searchResults.length} results`}
          </SearchSummary>
        </InputWrapper>
      </form>
      <SearchResultsTable data={searchResults} />
    </Container>
  );
};

export default Search;
