import React, { useMemo } from "react";
import { NiblPackage } from "../../services/niblData";
import { useTable, Column, useSortBy } from "react-table";
import { FaArrowDown, FaArrowUp, FaDownload } from "react-icons/fa";
import * as Table from "../../styles/table";
import { requestFile } from "../../services/animuxd";
import styled from "../../styles/styled";

const ActionsContainer = styled.div`
  display: flex;
  text-align: center;
  align-items: center;
  width: 100%;
  justify-content: center;
`;

type Props = {
  data: NiblPackage[];
};
const SearchResultsTable: React.FunctionComponent<Props> = ({ data }) => {
  const columns = useMemo<Column<NiblPackage>[]>(() => {
    return [
      {
        Header: "File",
        accessor: "name",
      },
      {
        Header: "Ep.",
        accessor: "episodeNumber",
      },
      {
        Header: "Size",
        accessor: "sizekbits",
        Cell: (cell) => {
          return cell.row.original.size;
        },
      },
      {
        id: "actions",
        Header: "",
        accessor: "botId",
        Cell: (cell) => {
          return (
            <ActionsContainer>
              <FaDownload
                data-testid="actionRequestFile"
                onClick={() => requestFile(cell.row.original)}
                style={{ cursor: "pointer" }}
              />
            </ActionsContainer>
          );
        },
      },
    ];
  }, []);

  const {
    getTableProps,
    getTableBodyProps,
    headers,
    rows,
    prepareRow,
  } = useTable<NiblPackage>(
    {
      columns,
      data,
    },
    useSortBy
  );

  return (
    <Table.Container>
      <Table.Table {...getTableProps()}>
        <thead>
          <tr>
            {headers.map((column) => (
              <Table.Header
                {...column.getHeaderProps(column.getSortByToggleProps())}
              >
                {column.render("Header")}
                &nbsp;
                {column.isSorted ? (
                  column.isSortedDesc ? (
                    <FaArrowDown fontSize="1rem" />
                  ) : (
                    <FaArrowUp fontSize="1rem" />
                  )
                ) : null}
              </Table.Header>
            ))}
          </tr>
        </thead>
        <tbody {...getTableBodyProps()}>
          {rows.map((row) => {
            prepareRow(row);
            return (
              <Table.Row {...row.getRowProps()} data-testid="searchResultsRow">
                {row.cells.map((cell) => {
                  return (
                    <Table.Data {...cell.getCellProps()}>
                      {cell.render("Cell")}
                    </Table.Data>
                  );
                })}
              </Table.Row>
            );
          })}
        </tbody>
      </Table.Table>
    </Table.Container>
  );
};

export default React.memo(SearchResultsTable);
