# frozen_string_literal: true
class KyouenPuzzle
  attr_accessor :id, :stage_no, :size, :stage, :creator

  def initialize(entity = nil)
    return if entity.nil?
    @id = entity.key.path[0].id
    @stage_no = entity.properties['stageNo'].integer_value.to_i
    @size = entity.properties['size'].integer_value.to_i
    @stage = entity.properties['stage'].string_value
    @creator = entity.properties['creator'].string_value
  end

  class << self
    def find_by_stage_no(stage_no)
      client = Datastore::Client.new
      result = client.query(
        'SELECT * FROM KyouenPuzzle WHERE stageNo = @1',
        [
          Datastore::Parameter.new(stage_no)
        ]
      )
      result.map { |r| KyouenPuzzle.new(r.entity) }.first if result.present?
    end

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
