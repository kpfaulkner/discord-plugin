import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface MyQuery extends DataQuery {
  queryText?: string;
  rgSplit?: string;
  constant: number;
}

export const defaultQuery: Partial<MyQuery> = {
  constant: 6.5,
};

/**
 * These are options configured for each Discord Guild
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  discordGuildID?: string;     // From Discord
  discordGuildIDKey?: string;  // Made up for this plugin.
}


