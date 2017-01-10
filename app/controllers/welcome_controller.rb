class WelcomeController < ApplicationController

  def index
    client = Datastore::Client.new
    result = client.query(
      'SELECT * FROM KyouenPuzzle ORDER BY stageNo LIMIT @1',
      [Datastore::Parameter.new(5)]
    )
    render json: result[0].entity.as_json
  end
end
