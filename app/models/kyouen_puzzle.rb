# frozen_string_literal: true
class KyouenPuzzle
  attr_accessor :id, :stage_no, :size, :stage, :creator

  def initialize(entity)
    @id = entity.key.path[0].id
    @stage_no = entity.properties['stageNo'].integer_value
    @size = entity.properties['size'].integer_value
    @stage = entity.properties['stage'].string_value
    @creator = entity.properties['creator'].string_value
  end

  class << self
    def fetch(offset, limit)
      client = Datastore::Client.new
      result = client.query(
        'SELECT * FROM KyouenPuzzle ORDER BY stageNo LIMIT @1 OFFSET @2',
        [
          Datastore::Parameter.new(limit),
          Datastore::Parameter.new(offset)
        ]
      )
      result.map { |r| KyouenPuzzle.new(r.entity) }
    end
  end
end
