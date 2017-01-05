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

    dummyFile = nil
    dummyFile = DummyFile.new(ENV['GCP_KEY']) if ENV['GCP_KEY'].present?
    datastore.authorization = ServiceAccountCredentials.make_creds(
        json_key_io: dummyFile || File.open('cert/my-android-server-91c8d931db89.json'),
        scope: [
          Datastore::AUTH_CLOUD_PLATFORM,
          Datastore::AUTH_DATASTORE
        ]
    )
    datastore.authorization.fetch_access_token!

    query = GqlQuery.new(
        query_string: 'SELECT * FROM GcmModel LIMIT @limit',
        named_bindings: {
          limit: GqlQueryParameter.new(value: Value.new(integer_value: '5'))
        }
    )
    request = RunQueryRequest.new(gql_query: query)
    result = datastore.run_project_query('my-android-server', request)
    print result.batch.entity_results[0].entity.as_json
    # render json: result.batch.entity_results[0].entity.as_json
    render json: result.batch.entity_results.size
  end

  class DummyFile
    attr_accessor :read
    def initialize(read)
       @read = read
    end
  end
end
