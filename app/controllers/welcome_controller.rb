require 'googleauth'
require 'google/apis/datastore_v1'

class WelcomeController < ApplicationController
  ServiceAccountCredentials  = Google::Auth::ServiceAccountCredentials
  Datastore  = Google::Apis::DatastoreV1
  RunQueryRequest = Google::Apis::DatastoreV1::RunQueryRequest
  GqlQuery = Google::Apis::DatastoreV1::GqlQuery
  GqlQueryParameter = Google::Apis::DatastoreV1::GqlQueryParameter
  Value = Google::Apis::DatastoreV1::Value

  def index
    datastore = Google::Apis::DatastoreV1::DatastoreService.new

    datastore.authorization = ServiceAccountCredentials.make_creds(
        json_key_io: File.open('cert/my-android-server-91c8d931db89.json'),
        scope: [
          Datastore::AUTH_CLOUD_PLATFORM,
          Datastore::AUTH_DATASTORE
        ]
    )
    datastore.authorization.fetch_access_token!

    query = GqlQuery.new(
        query_string: 'SELECT * FROM GcmModel LIMIT 5',
        allow_literals: true
    )
    request = RunQueryRequest.new(gql_query: query)
    result = datastore.run_project_query('my-android-server', request)
    print result.batch.entity_results[0].entity.as_json
    render json: result.batch.entity_results[0].entity.as_json
  end
end
