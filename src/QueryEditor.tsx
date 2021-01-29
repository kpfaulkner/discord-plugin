
import defaults from 'lodash/defaults';
import React, { ChangeEvent, PureComponent } from 'react';
import {LegacyForms} from '@grafana/ui';

import {QueryEditorProps} from '@grafana/data';
import { DataSource } from './DataSource';
import {defaultQuery,  MyDataSourceOptions, MyQuery } from './types';

const {  FormField  } = LegacyForms;

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;


export class QueryEditor extends PureComponent<Props> {
  onQueryTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, queryText: event.target.value });
  };

  onConstantChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, constant: parseFloat(event.target.value) });
    // executes the query
    onRunQuery();
  };

  onRGChange = (event: any) => {
    const { onChange, query } = this.props;
    onChange({ ...query, rgSplit: event.target.value });
  };

  onFieldValueChange = (event: any, _name?: string) => {
    const { onChange, query } = this.props;
    onChange({ ...query, rgSplit: event.target.value });
  };
  //private indvAnOutField: any;

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const { queryText, rgSplit} = query;
    //const { rgText } = query;
    return (
      <div className="gf-form">
        <FormField
          labelWidth={8}
          value={queryText || ''}
          onChange={this.onQueryTextChange}
          label="Subscription ID"
          tooltip="Subscription ID"
          inputWidth={100}
        />
        <select
          value={rgSplit}
          onChange={this.onRGChange}
        >
          <option value={'numusers'}>{'Total Number Of Users'}</option>
          <option value={'numjoined'}>{'Number of Users Joined'}</option>
          <option value={'numleft'}>{'Number of Users Left'}</option>
          <option value={'nummessages'}>{'Number of Messages'}</option>
        </select>

      </div>
    );
  }
}
