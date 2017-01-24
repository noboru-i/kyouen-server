# frozen_string_literal: true
class User
  attr_accessor :id, :user_id, :screen_name, :image, :clear_stage_count

  def initialize(entity)
    @id = entity.key.path[0].name
    @user_id = entity.properties['userId'].string_value
    @screen_name = entity.properties['screenName'].string_value
    @image = entity.properties['image'].string_value
    @clear_stage_count = entity.properties['clearStageCount']&.integer_value
    @api_token = entity.properties['apiToken']&.string_value
  end

  def generate_api_token
    @api_token = SecureRandom.hex
    client = Datastore::Client.new
    values = {
      'userId': Datastore::Parameter.new(user_id.to_s),
      'screenName': Datastore::Parameter.new(screen_name),
      'image': Datastore::Parameter.new(image.to_s),
      'apiToken': Datastore::Parameter.new(@api_token)
    }
    client.update(@id, values)
  end

  class << self
    def create_new_user(user_id, screen_name, image)
      client = Datastore::Client.new
      params = {
        'userId': Datastore::Parameter.new(user_id.to_s),
        'screenName': Datastore::Parameter.new(screen_name),
        'image': Datastore::Parameter.new(image.to_s)
      }
      client.insert(params, 'KEY' + user_id.to_s)
      find_by_user_id(user_id)
    end

    def find_by_user_id(user_id)
      client = Datastore::Client.new
      result = client.query(
        'SELECT * FROM User WHERE userId = @1',
        [
          Datastore::Parameter.new(user_id.to_s)
        ]
      )
      result.map { |r| User.new(r.entity) }.first if result.present?
    end
  end
end
